package parser

import (
	"reflect"
	"testing"

	"github.com/dmclink/flash-cli/internal/constant"
)

func TestIsFilter(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"Starts with +", args{"+foo"}, true},
		{"Starts with -", args{"-foo"}, true},
		{"Starts with :", args{":foo"}, true},
		{"Starts with 0", args{"0"}, true},
		{"Starts with 1", args{"1"}, true},
		{"Starts with 9", args{"9"}, true},
		{"id number range", args{"103-108"}, true},
		{"comma separated ids", args{"103,108,20"}, true},
		{"Invalid numbers", args{"14a"}, true}, // still resolve to true! this function doesnt check correctness, should fail ValidateFilter() check
		{"Ends with :", args{"foo:"}, true},
		{"Contains :", args{"foo:bar"}, true},
		{"Regular string", args{"Go is great!"}, false},
		{"Filter like string with space", args{"elements:au gold"}, false},
		// {"Contains delimiter", args{"elements:au::gold"}, false}, TODO: this will resolve to true in current implementation
		{"Regular command", args{"add"}, false},
		{"Command ending in number", args{"report3"}, false},
		{"Contains +", args{"foo+bar"}, false},
		{"Contains -", args{"flash-cli"}, false},
		{"standard flag starting with --", args{"--help"}, false},
		{"Valid UUID starts with number", args{"525ea494-4ef4-4208-a9b8-023207abb2c7"}, true},
		{"Valid UUID starts with letter", args{"b25ea494-4ef4-4208-a9b8-023207abb2c7"}, true},
		// resolves to true because it starts with a number but will fail the ValidateFilter() check
		{"Invalid UUID length starts with number", args{"525ea494-4ef4-4208-a9b-023207abb27"}, true},
		{"Invalid UUID character starts with number", args{"525xz494-4yf4-4208-a9b8-023207abb2c7"}, true},
		// resolves to false, rest of program treats it as a command instead of a filter
		{"Invalid UUID length starts with letter", args{"b25ea494-4ef4-4208-a9b-023207abb27"}, false},
		{"Invalid UUID character starts with letter", args{"b25xz494-4yf4-4208-a9b8-023207abb2c7"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsFilter(tt.args.s); got != tt.want {
				t.Errorf("IsFilter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateFilter(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"valid single number", args{"1"}, false},
		{"valid multi digit number", args{"10"}, false},
		{"valid range", args{"10-29"}, false},
		{"valid comma separated numbers", args{"10,29,30"}, false},
		{"valid comma separated with range", args{"10,29-30,39"}, false},
		{"negative number", args{"-10"}, false}, // registers as valid (no error) because it treats it as a "-" operator with the flag "10" rather than a negative id
		{"invalid number", args{"10e5"}, true},
		{"decimal number", args{"1.1"}, true},
		{"invalid number in range", args{"10,29a,30"}, true},
		{"unfinished range", args{"10-"}, true},
		{"unfinished comma separated list", args{"10,"}, true},
		{"Valid UUID starts with number", args{"525ea494-4ef4-4208-a9b8-023207abb2c7"}, false},
		{"Valid UUID starts with letter", args{"b25ea494-4ef4-4208-a9b8-023207abb2c7"}, false},
		{"Invalid UUID length starts with number", args{"525ea494-4ef4-4208-a9b-023207abb27"}, true},
		{"Invalid UUID character starts with number", args{"525xz494-4yf4-4208-a9b8-023207abb2c7"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateFilter(tt.args.s); (err != nil) != tt.wantErr {
				t.Errorf("ValidateFilter() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_reorder(t *testing.T) {
	type args struct {
		args []string
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		want1   int
		wantErr bool
	}{
		{
			"default command: no filters, commands or mods",
			args{[]string{"flash-cli"}},
			[]string{"flash-cli", constant.DEFAULT_COMMAND},
			2, false,
		},
		{
			"valid command: only command",
			args{[]string{"flash-cli", "review"}},
			[]string{"flash-cli", "review"},
			2, false,
		},
		{
			"valid command: with no filters and with mods",
			args{[]string{"flash-cli", "add", "this", "is::a", "flashcard"}},
			[]string{"flash-cli", "add", "this", "is::a", "flashcard"},
			2, false,
		},
		{
			"valid command: with single group filter and with mods",
			args{[]string{"flash-cli", "group:go", "add", "this", "is::a", "flashcard"}},
			[]string{"flash-cli", "add", "group:go", "this", "is::a", "flashcard"},
			3, false,
		},
		{
			"valid command: with multiple group filters and with mod",
			args{[]string{"flash-cli", "group:go", "group:programming", "add", "this", "is::a", "flashcard"}},
			[]string{"flash-cli", "add", "group:go", "group:programming", "this", "is::a", "flashcard"},
			4, false,
		},
		{
			"valid command: with multiple comma separated id filters",
			args{[]string{"flash-cli", "14,18,20", "review"}},
			[]string{"flash-cli", "review", "14,18,20"},
			3, false,
		},
		{
			"invalid id filter",
			args{[]string{"flash-cli", "14a", "review"}},
			[]string{"flash-cli", "14a", "review"},
			-1, true,
		},
		{
			"invalid multiple id filter",
			args{[]string{"flash-cli", "14,a,1", "review"}},
			[]string{"flash-cli", "14,a,1", "review"},
			-1, true,
		},
		{
			"invalid range id filter",
			args{[]string{"flash-cli", "14-a", "review"}},
			[]string{"flash-cli", "14-a", "review"},
			-1, true,
		},
		{
			"invalid range id filter with default command",
			args{[]string{"flash-cli", "14-a"}},
			[]string{"flash-cli", "14-a"},
			-1, true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := reorder(tt.args.args)
			if (err != nil) != tt.wantErr {
				t.Fatalf("reorder() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("reorder() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("reorder() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestValidateFilters(t *testing.T) {
	type args struct {
		filters []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"empty filters", args{[]string{}}, false},
		{"single valid", args{[]string{"1"}}, false},
		{"multiple valid", args{[]string{"1-10", "+linux", "group:foo"}}, false},
		{"single invalid", args{[]string{"+"}}, true},
		{"mixed invalid valid", args{[]string{"+foo", "10-", "-bar"}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateFilters(tt.args.filters); (err != nil) != tt.wantErr {
				t.Errorf("ValidateFilters() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCommandIdx(t *testing.T) {
	type args struct {
		args []string
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{"no command", args{[]string{"flash-cli"}}, -1},
		{"only command", args{[]string{"flash-cli", "review"}}, 1},
		{"command with filters", args{[]string{"flash-cli", "1-20", "25", "group:foo", "review"}}, 4},
		{"command with mods", args{[]string{"flash-cli", "add", "flashcard", "front::and", "back"}}, 1},
		{"command with filters and mods", args{[]string{"flash-cli", "group:foo", "group:bar", "add", "flashcard", "front::and", "back"}}, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CommandIdx(tt.args.args); got != tt.want {
				t.Errorf("CommandIdx() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindCommand(t *testing.T) {
	type args struct {
		args []string
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 int
	}{
		{"no command", args{[]string{"flash-cli"}}, "", -1},
		{"only command", args{[]string{"flash-cli", "review"}}, "review", 1},
		{"command with filters", args{[]string{"flash-cli", "1-20", "25", "group:foo", "review"}}, "review", 4},
		{"command with mods", args{[]string{"flash-cli", "add", "flashcard", "front::and", "back"}}, "add", 1},
		{"command with filters and mods", args{[]string{"flash-cli", "group:foo", "group:bar", "add", "flashcard", "front::and", "back"}}, "add", 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := FindCommand(tt.args.args)
			if got != tt.want {
				t.Errorf("FindCommand() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("FindCommand() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestParseArgs(t *testing.T) {
	type args struct {
		args []string
	}
	tests := []struct {
		name    string
		args    args
		want    ParsedArgs
		wantErr bool
	}{
		{
			"no command",
			args{[]string{"flash-cli"}},
			ParsedArgs{
				constant.DEFAULT_COMMAND,
				[]string{},
				[]string{},
				"flash-cli",
				nil,
			},
			false,
		},
		{
			"only command",
			args{[]string{"flash-cli", "summary"}},
			ParsedArgs{
				"summary",
				[]string{},
				[]string{},
				"flash-cli summary",
				nil,
			},
			false,
		},
		{
			"command with filters",
			args{[]string{"flash-cli", "1-20", "25", "group:foo", "review"}},
			ParsedArgs{
				"review",
				[]string{"1-20", "25", "group:foo"},
				[]string{},
				"flash-cli 1-20 25 group:foo review",
				nil,
			},
			false,
		},
		{
			"command with mods",
			args{[]string{"flash-cli", "add", "flashcard", "front::and", "back"}},
			ParsedArgs{
				"add",
				[]string{},
				[]string{"flashcard", "front::and", "back"},
				"flash-cli add flashcard front::and back",
				nil,
			},
			false,
		},
		{
			"command with filters and mods",
			args{[]string{"flash-cli", "group:foo", "group:bar", "add", "flashcard", "front::and", "back"}},
			ParsedArgs{
				"add",
				[]string{"group:foo", "group:bar"},
				[]string{"flashcard", "front::and", "back"},
				"flash-cli group:foo group:bar add flashcard front::and back",
				nil,
			},
			false,
		},
		{
			"malformed filters",
			args{[]string{"flash-cli", "group:foo", "14a", "group:bar", "add", "flashcard", "front::and", "back"}},
			ParsedArgs{},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseArgs(tt.args.args)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseArgs() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseArgs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParsedArgs_Args(t *testing.T) {
	type args struct {
		binaryName string
	}
	type fields struct {
		Command       string
		Filters       []string
		Mods          []string
		OriginalInput string
	}
	tests := []struct {
		name   string
		args   args
		fields fields
		want   []string
	}{
		{
			"no command",
			args{"flash-cli"},
			fields{Command: "review", Mods: []string{}, Filters: []string{}, OriginalInput: "flash-cli"},
			[]string{"flash-cli", "review"},
		},
		{
			"only command",
			args{"flash-cli"},
			fields{Command: "review", Mods: []string{}, Filters: []string{}, OriginalInput: "flash-cli review"},
			[]string{"flash-cli", "review"},
		},
		{
			"with mods",
			args{"flash-cli"},
			fields{Command: "add", Mods: []string{"some", "card::and", "back"}, Filters: []string{}, OriginalInput: "flash-cli add some card::and back"},
			[]string{"flash-cli", "add", "some", "card::and", "back"},
		},
		{
			"with filters and mods",
			args{"flash-cli"},
			fields{Command: "add", Mods: []string{"some", "card::and", "back"}, Filters: []string{"group:foo"}, OriginalInput: "flash-cli group:foo add some card::and back"},
			[]string{"flash-cli", "add", "group:foo", "some", "card::and", "back"},
		},
		{
			"with filters",
			args{"flash-cli"},
			fields{Command: "review", Mods: []string{}, Filters: []string{"group:foo"}, OriginalInput: "flash-cli group:foo review"},
			[]string{"flash-cli", "review", "group:foo"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := ParsedArgs{
				Command:       tt.fields.Command,
				Filters:       tt.fields.Filters,
				Mods:          tt.fields.Mods,
				OriginalInput: tt.fields.OriginalInput,
			}
			if got := args.Args(tt.args.binaryName); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParsedArgs.Args() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParsedArgs_parseFilters(t *testing.T) {
	type fields struct {
		Command       string
		Filters       []string
		Mods          []string
		OriginalInput string
		parsedFilters *[]Filter
	}
	want := []Filter{{GROUP, "group", "foo", false, -1, -1, "group:foo"}, {GROUP, "group", "bar", false, -1, -1, "group:bar"}}
	tests := []struct {
		name   string
		fields fields
	}{
		{"unparsed", fields{"review", []string{"group:foo,bar"}, []string{}, "flash-cli group:foo,bar review", nil}},
		{"already parsed", fields{"review", []string{"group:foo,bar"}, []string{}, "flash-cli group:foo,bar review", &want}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := &ParsedArgs{
				Command:       tt.fields.Command,
				Filters:       tt.fields.Filters,
				Mods:          tt.fields.Mods,
				OriginalInput: tt.fields.OriginalInput,
				parsedFilters: tt.fields.parsedFilters,
			}
			args.parseFilters()

			if got := *args.parsedFilters; !reflect.DeepEqual(got, want) {
				t.Errorf("ParseArgs.parseFilters() = %v, want %v", got, want)
			}
		})
	}
}
