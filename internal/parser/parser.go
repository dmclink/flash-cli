package parser

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/dmclink/flash-cli/internal/constant"
	"github.com/dmclink/flash-cli/internal/utils"
)

// ParsedArgs are the command line arguments separated into categories split at parsed command
type ParsedArgs struct {
	Binary  string
	Command string
	Filters []string
	Mods    []string
	// OriginalInput is the original command line input from os.Args
	// before parsing/reordering. Preserved mostly for debugging
	// since os.Args is overwritten after reordering
	OriginalInput string
}

// Args outputs the cobra CLI friendly reordered arguments in order of
// <program> <command> <filters> <mods>
func (args ParsedArgs) Args(binaryName string) []string {
	return slices.Concat([]string{binaryName, args.Command}, args.Filters, args.Mods)
}

// CobraArgs returns the args with mods and filters stripped in a cobra CLI friendly manner.
// The "help" command is different in which it will preserve one mod (the subcommand) if exists
func (args ParsedArgs) CobraArgs() []string {
	if args.Command == "help" && len(args.Mods) > 0 {
		return []string{args.Binary, args.Command, args.Mods[0]}
	}
	return []string{args.Binary, args.Command}
}

// ParseArgs returns a struct for the parsed command line arguments. Returns an error if filters are malformed.
// Does not check if filters or mods are relevant to the individual subcommands requirements'
// Takes valid command line arguments in the form <program> <filters> <command> <mods> and
// separates them into a field in their respective types. Also returns args reordered into
// Cobra cli expected format <program> <command> <args>
func ParseArgs(args []string, commands map[string]bool) (ParsedArgs, error) {
	cmd, idx := FindCommand(args, commands)

	// reorderedArgs, modsStartIdx, err := reorder(args)
	// if err != nil {
	// 	return ParsedArgs{}, err
	// }

	var result ParsedArgs
	if idx == -1 {
		result = ParsedArgs{
			Binary:        args[0],
			Command:       cmd,
			Filters:       args[1:],
			Mods:          make([]string, 0),
			OriginalInput: strings.Join(args, " "),
		}
	} else {
		result = ParsedArgs{
			Binary:        args[0],
			Command:       cmd,
			Filters:       args[1:idx],
			Mods:          args[idx+1:],
			OriginalInput: strings.Join(args, " "),
		}
	}
	fmt.Println("FILTERS:", result.Filters)

	err := ValidateFilters(result.Filters)
	if err != nil {
		return ParsedArgs{}, err
	}

	return result, nil
}

// ValidateFilters returns an error on the first invalid filter found
// Empty filters slice is considered valid
func ValidateFilters(filters []string) error {
	for _, filter := range filters {
		fmt.Println("VALIDATING: ", filter)
		if !IsFilter(filter) {
			return fmt.Errorf("not a filter")
		}

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

	if strings.HasPrefix(s, "-") && !strings.HasPrefix(s, "--") {
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
//
// Preconditions: s passed the IsFilter() function check
// Passing IsFilter() ensures one of these conditions:
//   - s starts with "-" but not "--" ie. "-foo"
//   - s starts with "+" ie. "+foo"
//   - s contains ":" ie. "group:foo" "foo:bar" ":"
//   - s starts with a digit "[0..9]" ie. "1,5,10" "14a"
//   - s is a valid UUID
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
			for idFilter := range strings.SplitSeq(idFilterAndRanges, "-") {
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
func CommandIdx(args []string, commands map[string]bool) int {
	// skip program name at index 0
	for i := 1; i < len(args); i++ {
		arg := args[i]
		if commands[arg] {
			return i
		}
	}

	return -1
}

// FindCommand returns the command and its index in a slice of args
// Returns default command and -1 if no command is found
// Preconditions: args is a full command line input from os.Args, including program name `flash-cli` at index 0
func FindCommand(args []string, commands map[string]bool) (string, int) {
	idx := CommandIdx(args, commands)
	if idx == -1 {
		return constant.DEFAULT_COMMAND, idx
	}

	return args[idx], idx
}
