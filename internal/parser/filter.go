package parser

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/dmclink/flash-cli/internal/utils"
	"github.com/spf13/viper"
)

type FilterType int

const (
	ID FilterType = iota
	RANGE
	UUID
	TAG
	GROUP
	KV
	CUSTOM
)

func unalias(key string) string {
	aliases := map[string]string{
		"grp":     "group",
		"groups":  "group",
		"project": "group",
		"proj":    "group",
		"prj":     "group",
		"lim":     "limit",
	}

	if a, ok := aliases[key]; ok {
		return a
	}
	return key
}

// SearchFtilers is an object to hold arrays of all filter types found in args
// separated into their respective buckets
type SearchFilters struct {
	IDs    []Filter
	Ranges []Filter
	UUIDs  []Filter
	Tags   []Filter
	Groups []Filter
	KVs    []Filter
}

func (sf *SearchFilters) Size() int {
	result := 0
	result += len(sf.IDs)
	result += len(sf.Ranges)
	result += len(sf.UUIDs)
	result += len(sf.Tags)
	result += len(sf.Groups)
	// skipping KVs here on purpose because it is not used to build database search queries

	return result
}

func (sf *SearchFilters) Limit() (int, error) {
	result := -1
	for _, f := range sf.KVs {
		if f.Key == "limit" {
			n, err := strconv.Atoi(f.Value)
			if err != nil {
				return -1, fmt.Errorf("invalid value '%s' passed to limit; must be a number", f.Value)
			}
			result = n
		}
	}
	return result, nil
}

func ParseSearchFilters(v *viper.Viper, parsedArgs ParsedArgs, usesDefaultFilters bool) SearchFilters {
	filters := ParseFilters(parsedArgs)
	return NewSearchFilters(v, filters, usesDefaultFilters)
}

func (sf SearchFilters) CanOverrideGroups() bool {
	return len(sf.IDs)+len(sf.UUIDs)+len(sf.Ranges)+len(sf.Groups) == 0
}

func (sf SearchFilters) CanOverrideTags() bool {
	return sf.Size() == 0
}

func splitFieldsAndCommas(s string) []string {
	intermediate := strings.Fields(s)
	result := []string{}
	for _, i := range intermediate {
		result = append(result, splitAtCommas(i)...)
	}
	return result
}

func overrideSetting(v *viper.Viper, cfgKey string, prefix string, targetSlice *[]Filter) {
	configString := v.GetString(cfgKey)
	if configString != "" {
		configs := splitFieldsAndCommas(configString)
		rfs := toRawFiltersWithPrefix(configs, prefix)
		newFilters := make([]Filter, 0, len(rfs))
		for _, rf := range rfs {
			newFilters = append(newFilters, rf.toFilter())
		}
		*targetSlice = append(*targetSlice, newFilters...)
	}
}

func (sf *SearchFilters) TryOverrideGroups(v *viper.Viper) {
	overrideSetting(v, "default.filter.groups", "group:", &sf.Groups)
}

func (sf *SearchFilters) TryOverrideTags(v *viper.Viper) {
	overrideSetting(v, "default.filter.tags", "+", &sf.Tags)
}

func NewSearchFilters(v *viper.Viper, filters []Filter, usesDefaultFilters bool) SearchFilters {
	result := SearchFilters{}
	for _, filter := range filters {
		typ := filter.Type
		switch typ {
		case ID:
			result.IDs = append(result.IDs, filter)
		case RANGE:
			result.Ranges = append(result.Ranges, filter)
		case UUID:
			result.UUIDs = append(result.UUIDs, filter)
		case GROUP:
			result.Groups = append(result.Groups, filter)
		case TAG:
			result.Tags = append(result.Tags, filter)
		case KV:
			result.KVs = append(result.KVs, filter)
			// CUSTOM type filters aren't currently included in the DB query building
			// I will likely handle them differently somewhere else, as such we don't want to increment size
			// or it will cause an empty WHERE clause to be built
			// If I instead include custom into the regular db search query then delete this line
		default:
			fmt.Println("unexpected type creating new search filter")
			panic("code error")
		}
	}

	if usesDefaultFilters {
		if result.CanOverrideGroups() {
			result.TryOverrideGroups(v)
		}

		if result.CanOverrideTags() {
			result.TryOverrideTags(v)
		}
	}

	return result
}

// Filter is a single filter (no compounds) which hold the string and the filter type
// Values stored in Type are one of constants ID|RANGE|UUID|TAG|GROUP|CUSTOM
type Filter struct {
	// Type is an iota enum one of ID|RANGE|UUID|TAG|GROUP|CUSTOM that represents the filter's type
	Type FilterType
	// Key is the filter key for CUSTOM type filters. Will equal "group" for GROUP types and
	// its respective +/- tag for TAGs though other fields should be used. All others will be empty string
	Key string
	// Value is filter string processed for database query. Strips prefixes
	Value string
	// True if it was a '-' tag. Will be global mandate exclusion rather than 'or' behavior.
	IsExclude bool
	// Low end of the ID range, inclusive. Will equal High if single ID. Will equal -1 for all other types.
	Low int
	// High end of the ID range, inclusive. Will equal Low if single ID. Will equal -1 for all other types.
	High int
	// f is the raw filter string
	f string
}

func (f Filter) String() string {
	return fmt.Sprintf("Filter{%v %s %s %v %v %v %s}", f.Type, f.Key, f.Value, f.IsExclude, f.Low, f.High, f.f)
}

// IsMandated returns true if filter is a TAG type, all other types return false by default.
// Mandated in this context refers to SQL queries for this filter to have implicit 'and' behavior
// rather than default 'or'.
func (f Filter) IsMandated() bool {
	return f.Type == TAG
}

// ParseFilters takes a ParsedArgs struct and extracts its filters field. It then separates all
// compound filters into a set of equivalent single filters with their respective filter types assigned.
// ID range filters (ie. "5-9") get additional parsing into a struct with their integer values
//
// NOTE: Original order of filters is not preserved
//
// Preconditions: Assumes all filters in args passed validation check by parser.ValidateFilter()
// regardless of type.
func ParseFilters(args ParsedArgs) []Filter {
	// cast all raw filter strings to RawFilter type to provide parsing methods
	rawFilters := make([]RawFilter, 0, len(args.Filters))
	for _, f := range args.Filters {
		rawFilters = append(rawFilters, RawFilter(f))
	}

	// separate compound filters so they can first be broken down into their individual parts
	compoundFilters := make([]RawFilter, 0, len(args.Filters))
	singleRawFilters := []RawFilter{}
	for _, f := range rawFilters {
		if f.isCompound() {
			compoundFilters = append(compoundFilters, f)
		} else {
			singleRawFilters = append(singleRawFilters, f)
		}
	}

	// break apart and rebuild the compound filters
	// redundant getType parsing here and building unique interfaces. the split function for each needs to be different
	// enough to be annoying to make a generic type for it, performance shouldnt be a big hit
	compoundTypedFilters := make([]compoundTypedFilter, 0, len(compoundFilters))
	for _, f := range compoundFilters {
		// could pass down the type into a second array or new interface or something to avoid double parsing
		// getType (it gets called again later in toFilter() after they're split into individual filters)
		// but I'm lazy and parsing is simple enough it's probably not worth added complexity
		switch f.getType() {
		case KV:
			compoundTypedFilters = append(compoundTypedFilters, compoundCustomFilter{baseFilter{f}})
		case ID:
			compoundTypedFilters = append(compoundTypedFilters, compoundIDFilter{baseFilter{f}})
		case TAG:
			compoundTypedFilters = append(compoundTypedFilters, compoundTagFilter{baseFilter{f}})
		default:
			panic("untyped filter, should never reach here")
		}
	}
	for _, f := range compoundTypedFilters {
		singleRawFilters = append(singleRawFilters, f.split()...)
	}

	result := make([]Filter, 0, len(singleRawFilters))
	for _, rf := range singleRawFilters {
		f := rf.toFilter()
		result = append(result, f)
	}

	return result
}

// toFilter parses the RawFilter to determine its type then converts it into a Filter
func (rf RawFilter) toFilter() Filter {
	typ := rf.getType()
	s := rf.String()
	switch typ {
	case ID:
		id, _ := strconv.Atoi(s)
		return Filter{typ, "", s, false, id, id, s}
	case RANGE:
		sp := strings.Split(s, "-")
		lo, err := strconv.Atoi(sp[0])
		if err != nil {
			fmt.Println("invalid range passed to toFilter. this should have been processsed and casted correctly | filter: ", rf.String())
			panic("code error")
		}
		hi, err := strconv.Atoi(sp[1])
		if err != nil {
			fmt.Println("invalid range passed to toFilter. this should have been processsed and casted correctly | filter: ", rf.String())
			panic("code error")
		}
		if hi < lo {
			lo, hi = hi, lo
		}
		return Filter{typ, "", s, false, lo, hi, s}
	case UUID:
		return Filter{typ, "", s, false, -1, -1, s}
	case TAG:
		isExcl := s[0] == '-'
		return Filter{typ, s[0:1], s[1:], isExcl, -1, -1, s}
	case KV:
		sp := strings.SplitN(s, ":", 2)
		if len(sp) != 2 {
			sp = strings.SplitN(s, "=", 2)
		}
		k := unalias(sp[0])
		if k == "group" {
			return Filter{GROUP, k, sp[1], false, -1, -1, k + ":" + sp[1]}
		}
		return Filter{typ, k, sp[1], false, -1, -1, k + ":" + sp[1]}
	default:
		panic("unexpected type in RawFilter, likely code error")
	}
}

// baseFilter holds the RawFilter field and provides a string method to print it back out as a string
type baseFilter struct {
	f RawFilter
}

// String prints the RawFilter as a string
func (b baseFilter) String() string {
	return string(b.f)
}

// RawFilter is a command line filter in string based form of any type
//
// Precondition: any value converted to RawFilter must already passed parser.ValidateFilter() check
// and is at least one valid filter type. Passing an invalid filter will result in incorrect behavior
type RawFilter string

func (rf RawFilter) String() string {
	return string(rf)
}

// compoundTypedFilter is a generic typed filter that provides a split method to break apart the
// compound filter into invidiual filters of the same type
type compoundTypedFilter interface {
	split() []RawFilter
}

type compoundGroupFilter struct {
	baseFilter
}

type compoundIDFilter struct {
	baseFilter
}

type compoundTagFilter struct {
	baseFilter
}

type compoundCustomFilter struct {
	baseFilter
}

// splitAtCommas delimits a string by commas and removes duplicate values
//
// NOTE: order is not maintained
func splitAtCommas(s string) []string {
	sp := strings.Split(s, ",")

	m := make(map[string]bool, len(sp))
	for _, ss := range sp {
		m[ss] = true
	}

	result := make([]string, 0, len(m))
	for k := range m {
		result = append(result, k)
	}

	return result
}

// toRawFiltersWithPrefix adds prefix to each string in filters and casts them into RawFilters
func toRawFiltersWithPrefix(filters []string, prefix string) []RawFilter {
	result := make([]RawFilter, 0, len(filters))
	for _, f := range filters {
		result = append(result, RawFilter(prefix+f))
	}

	return result
}

// split sheds the original "group:" or alias prefix, splits at commas, and outputs each individually with "group:" prefix
//
// NOTE: order is not maintained, original prefix is not maintained
func (f compoundGroupFilter) split() []RawFilter {
	sp := strings.Split(f.String(), ":")
	prefix := "group:" // lose aliases ie. "grp" "project" here but it shouldnt matter
	filters := splitAtCommas(sp[1])

	return toRawFiltersWithPrefix(filters, prefix)
}

// split splits the filter at the commas and outputs as a slice of individual filters
// does not add a prefix, unlike the other split() methods
//
// NOTE: order is not maintained
func (f compoundIDFilter) split() []RawFilter {
	return toRawFiltersWithPrefix(splitAtCommas(f.String()), "")
}

// split maintains the tag prefix ('+' or '-') and splits the tags at the commas.
// it outputs each as a slice of individual tags with the original tag
//
// NOTE: order is not maintained
func (f compoundTagFilter) split() []RawFilter {
	prefix := f.String()[:1]
	s := f.String()[1:]

	filters := splitAtCommas(s)

	return toRawFiltersWithPrefix(filters, prefix)
}

// split splits the custom filters by their commas and outputs a slice of individual
// filters each prefixed by the original custom name
// ie. foo:bar,baz becomes []{foo:bar, foo:baz}
//
// NOTE: order is not maintained
func (f compoundCustomFilter) split() []RawFilter {
	sp := strings.Split(f.String(), ":")
	prefix := sp[0] + ":"
	filters := splitAtCommas(sp[1])

	return toRawFiltersWithPrefix(filters, prefix)
}

// getType parses the filter to determine which FilterType it is.
// Preconditions: assumes any RawFilter has been validated with parser.ValidateFilter first
// otherwise may result in incorrect type returned or panic. Incorrect types may also cause
// panics further in the program as they rely on splits to return the correct length slices
func (filter RawFilter) getType() FilterType {
	// order matters here as nothing prevents tags from including ":" character
	// and isCustomDataType will always return true if isGroupType is true
	// UUID must be called before ID type as UUIDs can start with a digit
	switch {
	case filter.isTagType():
		return TAG
	case filter.isKVType():
		return KV
	case filter.isUUIDType():
		return UUID
	case filter.isRangeType():
		return RANGE
	case filter.isIDType():
		return ID
	default:
		panic("unknown filter type: " + filter.String())
	}
}

// isCompound returns true if filter contains any commas
func (filter RawFilter) isCompound() bool {
	return strings.Contains(filter.String(), ",")
}

// isTagType returns true if filter starts with + or -
func (filter RawFilter) isTagType() bool {
	return strings.HasPrefix(string(filter), "-") || strings.HasPrefix(string(filter), "+")
}

// isKVType returns true if filter contains any ":"
func (filter RawFilter) isKVType() bool {
	return strings.Contains(string(filter), ":") || strings.Contains(string(filter), "=")
}

// isUUIDType returns true if filter passes UUID validation
// NOTE: must be called before isIDType in switch and if/else blocks
// as UUIDs may start with a digit
func (filter RawFilter) isUUIDType() bool {
	return utils.IsValidUUID(filter.String())
}

// isRangeType returns true if the entire string matches <integer>-<integer> where integers are non-empty
// Though an ID is a range, a range is not an ID. Nor are compound ranges considered a "RANGE" type,
// though they will also pass the isIDType check. This behavior is intentional as this method is only
// intended to be used on singular filter types.
// panics only if regex pattern cannot be parsed singalling a code error
// NOTE: must be called before isIDType in switch or if/else blocks
// as isIDType will always be true if isRangeType is true
// Will never result in true for compound types
func (filter RawFilter) isRangeType() bool {
	pattern := regexp.MustCompile(`^\d+-\d+$`)
	return pattern.MatchString(filter.String())
}

// isIDType returns true if the first character in filter is a numerical digit
func (filter RawFilter) isIDType() bool {
	f := string(filter)
	return f[0] >= '0' && f[0] <= '9'
}
