package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/dmclink/flash-cli/internal/constant"
	"github.com/dmclink/flash-cli/internal/utils"
	"github.com/spf13/cobra"
)

//TODO: implement filter type check and assignment for clarity on code branching for validation
// type filterType int
//
// const (
// 	ID filterType = iota
// 	ID_RANGE
// 	MOD
// )

// ParsedArgs are the command line arguments separated into their respective types.
type ParsedArgs struct {
	Command string
	Filters []string
	Mods    []string
	// Reordered command line arguments in format <program> <command> <filters> <mods>
	Args []string
}

// ExtractParsedArgs extracts the ParsedArgs struct from the cmd's context
// Only returns error on failing to cast the contents of the context into the ParsedArgs struct
// which should not occur if added into context correctly
func ExtractParsedArgs(cmd *cobra.Command) (ParsedArgs, error) {
	parsedArgs, ok := cmd.Context().Value(constant.PARSED_ARGS_KEY).(ParsedArgs)
	if !ok {
		return ParsedArgs{}, fmt.Errorf("failed to cast ParsedArgs")
	}

	return parsedArgs, nil
}

// ParseArgs returns a struct for the parsed command line arguments. Returns an error if filters are malformed.
// Does not check if filters or mods are relevant to the individual subcommands requirements'
// Takes valid command line arguments in the form <program> <filters> <command> <mods> and
// separates them into a field in their respective types. Also returns args reordered into
// Cobra cli expected format <program> <command> <args>
func ParseArgs(args []string) (ParsedArgs, error) {
	reorderedArgs, modsStartIdx, err := Reorder(args)
	if err != nil {
		return ParsedArgs{}, err
	}

	result := ParsedArgs{
		Command: reorderedArgs[1],
		Filters: reorderedArgs[2:modsStartIdx],
		Mods:    reorderedArgs[modsStartIdx:],
		Args:    reorderedArgs,
	}

	return result, nil
}

// Reorder takes a slice of command line arguments (ie. os.Args) and reorders and returns them
// Input in the original format of <program> <filters> <command> <mods> to
// match Cobra CLIs expected format <program> <command> <args>.
//
// Returns the starting index of <mods> in the reordered slice since Cobra Cli
// does not differentiate between <filters> and <mods>.
//
// If no command is entered and input matches form <program> <filters>,
// it will insert the default command `review` and return length of args as index effectively
// denoting all arguments as filters
//
// Returns an error if any filters are malformed. On errors, returns args without reordering and index -1
func Reorder(args []string) ([]string, int, error) {
	cmd, idx := FindCommand(args)
	if idx == -1 {
		// no command found:
		// inject default command after filters as if user had entered it in the expected position
		return Reorder(append(append([]string{args[0]}, args[1:]...), constant.DEFAULT_COMMAND))
	}

	if idx == 1 {
		return args, 2, nil
	}

	newArgs := make([]string, len(args))
	newArgs[0] = args[0] // program name ie. `flash-cli`
	newArgs[1] = cmd     // command name ie. `add` put in front for cobra cli to recognize it

	filters := args[1:idx]
	if len(filters) > 0 {
		copy(newArgs[2:2+len(filters)], filters)
	}

	mods := args[idx+1:]
	if len(mods) > 0 {
		copy(newArgs[idx+1:], mods)
	}

	err := ValidateFilters(filters)
	if err != nil {
		return args, -1, fmt.Errorf("one or more invalid filters | %w", err)
	}

	return newArgs, idx + 1, nil
}

// ValidateFilters returns an error on the first invalid filter found
// Empty filters slice is considered valid
// Preconditions: all filters passed the IsFilter() function check
func ValidateFilters(filters []string) error {
	for _, filter := range filters {
		err := ValidateFilter(filter)
		if err != nil {
			return err
		}
	}

	return nil
}

// IsFilter produces true if the string matches the requirements of command filter.
// If a string contains any whitespace charcater it is not a filter
// Any single word string starting with '-' or '+' and any strings containing ':' are considered filters.
// Any string consisting of a valid uuid is considered a filter.
// Any string starting with a numeric digit is considered a filter.
//
// NOTE: This does not validate filter's correctness which must be checked elsewhere.
// ie. a filter character without arguments ":" will pass or invalid number will pass "14a"
func IsFilter(s string) bool {
	if strings.ContainsAny(s, " \n\t\r") {
		return false
	}

	// TODO: need to check if contains delimiter, need to implement delimiter from config first

	if strings.HasPrefix(s, "-") {
		return true
	}

	if strings.HasPrefix(s, "+") {
		return true
	}

	if strings.Contains(s, ":") {
		return true
	}

	if s[0] >= '0' && s[0] <= '9' {
		return true
	}

	if utils.IsValidUUID(s) {
		return true
	}

	return false
}

// ValidateFilter returns an error if filters passed to this function are invalid
// Preconditions: s passed the IsFilter() function check
func ValidateFilter(s string) error {
	if s == "+" || s == "-" {
		return fmt.Errorf("Invalid filter: needs text after +/- modifier")
	}

	// catch uuids before assuming filters starting with a digit are an id filter
	if utils.IsValidUUID(s) {
		return nil
	}

	// starts with digit, must be an integer, a range of integers, or a comma separated list of either
	// tries to cast all individual digit components into an int and throws error if any fail
	if s[0] >= '0' && s[0] <= '9' {
		for idFilterAndRanges := range strings.SplitSeq(s, ",") {
			// fmt.Println(idFilterAndRanges)
			for idFilter := range strings.SplitSeq(idFilterAndRanges, "-") {
				// fmt.Println(idFilter)
				_, err := strconv.Atoi(idFilter)
				if err != nil {
					return fmt.Errorf("Invalid numerical filter")
				}
			}
		}
	}

	return nil
}

// CommandIdx finds the index of the command in a slice of args.
// If no commands are found (all args are filters), then -1 is returned.
// Preconditions: args is a full command line input from os.Args, including program name `flash-cli` at index 0
func CommandIdx(args []string) int {
	// skip program name at index 0
	for i := 1; i < len(args); i++ {
		arg := args[i]
		if !IsFilter(arg) {
			return i
		}
	}

	return -1
}

// FindCommand returns the command and its index in a slice of args
// Returns empty string and -1 if no command is found
// Preconditions: args is a full command line input from os.Args, including program name `flash-cli` at index 0
func FindCommand(args []string) (string, int) {
	idx := CommandIdx(args)
	if idx == -1 {
		return "", idx
	}

	return args[idx], idx
}
