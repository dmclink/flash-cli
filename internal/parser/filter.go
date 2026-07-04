package parser

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/dmclink/flash-cli/internal/utils"
)

type FilterType int

const (
	ID FilterType = iota
	RANGE
	UUID
	TAG
	GROUP
	CUSTOM
)

// Filter is a single filter (no compounds) which hold the string and the filter type
// Values stored in Type are one of constants ID|RANGE|UUID|TAG|GROUP|CUSTOM
type Filter struct {
	// f is the raw filter string
	f string
	// Type is an iota enum one of ID|RANGE|UUID|TAG|GROUP|CUSTOM that represents the filter's type
	Type FilterType
	// Range is an optional field that exists for all RANGE filter types otherwise is nil
	// It holds the conversion of the raw filter type into integers
	Range *Range
}

func (f Filter) String() string {
	return f.f
}

// Range is a filter type that is an inclusive range of IDs.
// Low will always be the smaller of the two numbers
type Range struct {
	Low  int
	High int
}

// NewRange creates a new Range from the filter passed to it.
// Casts the string representations of the number on each side of the range to an int
// Ensures the smaller of the two is assigned to Low regardless of position in string
// Preconditions: only valid Range type filters are allowed passed to f.
// ie. they match the pattern <digit[s]>-<digit[s]>
// Both numbers are small enough to fit into an int
func NewRange(f RawFilter) Range {
	s := f.String()
	if f.getType() != RANGE {
		panic("this function is only to be called with RANGE type filters")
	}
	split := strings.Split(s, "-")
	lo, err := strconv.Atoi(split[0])
	if err != nil {
		panic("failed splitting range string. strings passed to this func should be format <int>-<int> | " + err.Error())
	}
	hi, err := strconv.Atoi(split[1])
	if err != nil {
		panic("failed splitting range string. strings passed to this func should be format <int>-<int> | " + err.Error())
	}

	if hi < lo {
		lo, hi = hi, lo
	}

	return Range{lo, hi}
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
		switch f.getType() {
		case GROUP:
			compoundTypedFilters = append(compoundTypedFilters, compoundGroupFilter{baseFilter{f}})
		case CUSTOM:
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
	for _, f := range singleRawFilters {
		if f.getType() == RANGE {
			r := NewRange(f)
			result = append(result, Filter{f.String(), f.getType(), &r})
		} else {
			result = append(result, Filter{f.String(), f.getType(), nil})
		}
	}

	return result
}

// GetGroups returns all GROUP type elements in filters and strips their "group:" prefixes
func GetGroups(filters []Filter) []string {
	result := make([]string, 0, len(filters))
	for _, f := range filters {
		if f.Type == GROUP {
			result = append(result, strings.TrimPrefix(f.String(), "group:"))
		}
	}

	return result
}

// GetTags returns all TAG type elements in filters
func GetTags(filters []Filter) []string {
	result := make([]string, 0, len(filters))
	for _, f := range filters {
		if f.Type == TAG {
			result = append(result, f.String())
		}
	}

	return result
}

// groupPrefixes are reserved prefixes to denote a GROUP filter. Any filter with a semi-colon not
// in this list of prefixes will be considered CUSTOM
var groupPrefixes = []string{"group:", "grp:", "project:", "proj:", "prj:", "groups:"}

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
	case filter.isGroupType():
		return GROUP
	case filter.isCustomDataType():
		return CUSTOM
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

// isGroupType returns true if filter starts with "group:" or one of its aliases
func (filter RawFilter) isGroupType() bool {
	f := string(filter)
	for _, prefix := range groupPrefixes {
		if strings.HasPrefix(f, prefix) {
			return true
		}
	}

	return false
}

// isTagType returns true if filter starts with + or -
func (filter RawFilter) isTagType() bool {
	return strings.HasPrefix(string(filter), "-") || strings.HasPrefix(string(filter), "+")
}

// isCustomDataType returns true if filter contains any ":"
// NOTE: Run this after isGroupType if using in a switch case or if/else block as both will be true
func (filter RawFilter) isCustomDataType() bool {
	return strings.Contains(string(filter), ":")
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
