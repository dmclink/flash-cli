package edit

import (
	"maps"
	"reflect"
	"slices"
	"strings"
	"testing"
)

func Test_stripCommentsFromLines(t *testing.T) {
	type args struct {
		lines []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			"template string",
			args{lines: strings.Split(template, "\n")},
			[]string{
				"  Groups:            %s",
				"  Tags:              %s",
				"  Created:           %s",
				"  Last Review:       %s",
				"  Front:             %s",
				"  Back:              %s",
				"  Ext Data:          %s",
				"  Quit Batch Edit:   false",
			},
		},
		{"comment line", args{lines: []string{"# Foobar"}}, []string{}},
		{"comment line not first column", args{lines: []string{" # Foobar"}}, []string{" # Foobar"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := stripCommentsFromLines(tt.args.lines); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("stripCommentsFromLines() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_stripBlankBeforeFirstName(t *testing.T) {
	type args struct {
		lines []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			"no blanks",
			args{lines: []string{
				"  Groups:            %s",
				"  Tags:              %s",
				"  Created:           %s",
				"  Last Review:       %s",
				"  Front:             %s",
				"  Back:              %s",
				"  Ext Data:          %s",
				"  Quit Batch Edit:   false",
			}},
			[]string{
				"  Groups:            %s",
				"  Tags:              %s",
				"  Created:           %s",
				"  Last Review:       %s",
				"  Front:             %s",
				"  Back:              %s",
				"  Ext Data:          %s",
				"  Quit Batch Edit:   false",
			},
		},
		{
			"blanks after front",
			args{lines: []string{
				"  Groups:            %s",
				"  Tags:              %s",
				"  Created:           %s",
				"  Last Review:       %s",
				"  Front:             %s",
				"",
				"  Back:              %s",
				"  Ext Data:          %s",
				"  Quit Batch Edit:   false",
			}},
			[]string{
				"  Groups:            %s",
				"  Tags:              %s",
				"  Created:           %s",
				"  Last Review:       %s",
				"  Front:             %s",
				"",
				"  Back:              %s",
				"  Ext Data:          %s",
				"  Quit Batch Edit:   false",
			},
		},
		{
			"no length lines before",
			args{lines: []string{
				"",
				"",
				"  Groups:            %s",
				"  Tags:              %s",
				"  Created:           %s",
				"  Last Review:       %s",
				"  Front:             %s",
				"  Back:              %s",
				"  Ext Data:          %s",
				"  Quit Batch Edit:   false",
			}},
			[]string{
				"  Groups:            %s",
				"  Tags:              %s",
				"  Created:           %s",
				"  Last Review:       %s",
				"  Front:             %s",
				"  Back:              %s",
				"  Ext Data:          %s",
				"  Quit Batch Edit:   false",
			},
		},
		{
			"only whitespace lines before",
			args{lines: []string{
				"   ",
				"\t",
				"  Groups:            %s",
				"  Tags:              %s",
				"  Created:           %s",
				"  Last Review:       %s",
				"  Front:             %s",
				"  Back:              %s",
				"  Ext Data:          %s",
				"  Quit Batch Edit:   false",
			}},
			[]string{
				"  Groups:            %s",
				"  Tags:              %s",
				"  Created:           %s",
				"  Last Review:       %s",
				"  Front:             %s",
				"  Back:              %s",
				"  Ext Data:          %s",
				"  Quit Batch Edit:   false",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := stripBlankBeforeFirstName(tt.args.lines); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("stripBlankBeforeFirstName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getNameFromLine(t *testing.T) {
	type args struct {
		line string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"valid", args{"  Groups:            %s"}, "Groups"},
		{"valid with extra semicolon", args{"  Groups::            %s"}, "Groups"},
		{"valid no name", args{"                     %s"}, ""},
		{"empty line", args{""}, ""},
		{"long whitespace line", args{"\t\n                                        \t\n"}, ""},
		{"invalid prefix whitespace", args{"Groups:              %s"}, "oups"},
		{"invalid post spacing", args{"  Groups:          %s"}, "Groups:          %s"},
		{"invalid comment line", args{"# Groups:            %s"}, "Groups"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getNameFromLine(tt.args.line); got != tt.want {
				t.Errorf("getNameFromLine() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_splitLine(t *testing.T) {
	type args struct {
		line string
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 string
	}{
		{"valid", args{"  Groups:            %s"}, "Groups", "%s"},
		{"valid no name", args{"                     %s"}, "", "%s"},
		{"valid no data", args{"  Groups:              "}, "Groups", ""},
		{"empty line", args{""}, "", ""},
		{"invalid comment line", args{"# ID:                1"}, "ID", "1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := splitLine(tt.args.line)
			if got != tt.want {
				t.Errorf("splitLine() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("splitLine() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_getValueFromLine(t *testing.T) {
	type args struct {
		line string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"valid", args{"  Groups:            %s"}, "%s"},
		{"valid no name", args{"                     %s"}, "%s"},
		{"empty line", args{""}, ""},
		{"long whitespace line", args{"\t\n                                        \t\n"}, ""},
		{"tab indents before text", args{"                     \tfoo"}, "\tfoo"},
		{"whitespace indents before text", args{"                         foo"}, "    foo"},
		{"whitespace after text", args{"                     foo     "}, "foo"},
		{"invalid prefix whitespace", args{"Groups:              %s"}, "%s"},
		{"invalid post spacing", args{"  Groups:          %s"}, ""},
		{"invalid comment line", args{"# Groups:            %s"}, "%s"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getValueFromLine(tt.args.line); got != tt.want {
				t.Errorf("getValueFromLine() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_extractFields(t *testing.T) {
	exampleFullInput := `# The 'flash-cli <id> edit' command allows you to modify all aspects of a card
# using a text editor. Below is a representation of all the task details.
# Modify what you wish, and when you save and quit your editor,
# The program will read this file, determine what changed, and apply
# those changes. If you exit your editor without saving or making
# modifications, no changes will be made.
#
# Lines that begin with # represent data you cannot change, like ID.
# Edits to these lines will be ignored. The program will attempt to detect
# malformed edits and re-open the file with original content and an error 
# message displayed, no guarantees though. If you get stuck in a loop
# with the same file, just quit without saving.
#
# Do not reorder rows. Entered data must align, only data from 22nd column 
# and after is parsed. Do not edit row Name.
#
# Name               Editable details
# -----------------  ----------------------------------------------------
# ID:                1
# UUID:              some
# Editing Error:     Hey you messed up last time
#                    Don't do it again
# -----------------  ----------------------------------------------------
# Separate the tags and groups with spaces like this: tag1 tag2
# Group and tag names must start with a letter and not contain commas ','
  Groups:            programming java
  Tags:              hard needsreview quiz
  Created:           2006-01-02 15:04:05
  Last Review:       2026-07-30 12:00:01
# For multiline data for Front and Back, new lines must start on the same column
# That is they must be preceded by at least 21 whitespaces, careful not to use tabs!
  Front:             This is a card front! Pick the best one:
                     a. foo
                     b. bar
                     c. baz
  Back:              The answer is always 'c'.
# Below is the raw JSON data. Edit carefully. Must remain valid JSON.
  Ext Data:          {
                       "sm2": {
                         "difficulty": "hard"
                       }
                     }
# Flip this to true, save, and quit the editor to exit the batch editing loop early
# If true, skips applying any updates for this card.
  Quit Batch Edit:   false
# END`
	type args struct {
		data []byte
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{"valid", args{
			[]byte(exampleFullInput),
		}, map[string]string{
			"Groups":          "programming java",
			"Tags":            "hard needsreview quiz",
			"Created":         "2006-01-02 15:04:05",
			"Last Review":     "2026-07-30 12:00:01",
			"Front":           "This is a card front! Pick the best one:\na. foo\nb. bar\nc. baz",
			"Back":            "The answer is always 'c'.",
			"Ext Data":        "{\n  \"sm2\": {\n    \"difficulty\": \"hard\"\n  }\n}",
			"Quit Batch Edit": "false",
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractFields(tt.args.data)

			gotKeys := slices.Collect(maps.Keys(got))
			wantKeys := slices.Collect(maps.Keys(tt.want))
			slices.Sort(gotKeys)
			slices.Sort(wantKeys)
			if !reflect.DeepEqual(gotKeys, wantKeys) {
				t.Errorf("extractFields() incorrect map keys = %v, want %v", gotKeys, wantKeys)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("extractFields() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_toEditorInput(t *testing.T) {
	type args struct {
		fields map[string]string
	}
	tests := []struct {
		name    string
		args    args
		want    EditorInput
		wantErr bool
	}{
		{
			"valid",
			args{
				fields: map[string]string{
					"Groups":          "programming java",
					"Tags":            "hard needsreview quiz",
					"Created":         "2006-01-02 15:04:05",
					"Last Review":     "2026-07-30 12:00:01",
					"Front":           "This is a card front! Pick the best one:\na. foo\nb. bar\nc. baz",
					"Back":            "The answer is always 'c'.",
					"Ext Data":        "{\n  \"sm2\": {\n    \"difficulty\": \"hard\"\n  }\n}",
					"Quit Batch Edit": "false",
				},
			},
			EditorInput{
				Groups:        []string{"programming", "java"},
				Tags:          []string{"hard", "needsreview", "quiz"},
				CreatedAt:     1136185445,
				LastReview:    1785384001,
				Front:         "This is a card front! Pick the best one:\na. foo\nb. bar\nc. baz",
				Back:          "The answer is always 'c'.",
				ExtData:       []byte("{\n  \"sm2\": {\n    \"difficulty\": \"hard\"\n  }\n}"),
				QuitBatchEdit: false,
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := toEditorInput(tt.args.fields)
			if (err != nil) != tt.wantErr {
				t.Fatalf("toEditorInput() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("toEditorInput() = %v, want %v", got, tt.want)
			}
		})
	}
}
