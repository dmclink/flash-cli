package database

import (
	"database/sql"
	"fmt"
	"reflect"
	"testing"

	"github.com/dmclink/flash-cli/internal/constant"
	"github.com/dmclink/flash-cli/internal/parser"
	"github.com/google/go-cmp/cmp"
)

func TestAddFlashcard(t *testing.T) {
	type args struct {
		db     *sql.DB
		front  string
		back   string
		groups []parser.Filter
		tags   []parser.Filter
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := AddFlashcard(tt.args.db, tt.args.front, tt.args.back, tt.args.groups, tt.args.tags); (err != nil) != tt.wantErr {
				t.Errorf("AddFlashcard() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetAllFlashcards(t *testing.T) {
	type args struct {
		db *sql.DB
	}
	tests := []struct {
		name    string
		args    args
		want    []Flashcard
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetAllFlashcards(tt.args.db)
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetAllFlashcards() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAllFlashcards() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetFlashcards(t *testing.T) {
	type args struct {
		db      *sql.DB
		filters parser.SearchFilters
	}
	tests := []struct {
		name    string
		args    args
		want    []Flashcard
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetFlashcards(tt.args.db, tt.args.filters)
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetFlashcards() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetFlashcards() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_buildFlashcardSelectQuery(t *testing.T) {
	type args struct {
		filters parser.SearchFilters
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 []any
	}{
		{
			"no filters",
			args{parser.SearchFilters{Size: 0}},
			fmt.Sprintf("SELECT\n\tf.id, f.uuid, f.last_review, f.front, f.back, f.created_at, f.ext_data\nFROM\n\t%s f;", constant.DATABASE_TABLE_FLASHCARDS),
			[]any{},
		},
		{
			"id filter",
			args{parser.SearchFilters{Size: 1, IDs: []parser.Filter{
				{
					Type:      parser.ID,
					Key:       "",
					Value:     "1",
					IsExclude: false,
					Low:       1,
					High:      1,
				},
			}}},
			"SELECT\n\tf.id, f.uuid, f.last_review, f.front, f.back, f.created_at, f.ext_data\nFROM\n\t" + constant.DATABASE_TABLE_FLASHCARDS + " f\n" +
				"WHERE\n" +
				"\t(\n" +
				"\t\tf.id = ?\n" +
				"\t);",
			[]any{1},
		},
		{
			"uuid filter",
			args{parser.SearchFilters{Size: 1, UUIDs: []parser.Filter{
				{
					Type:      parser.UUID,
					Key:       "",
					Value:     "e3e9df30-bc8a-4458-af31-18fd437342fd",
					IsExclude: false,
					Low:       -1,
					High:      -1,
				},
			}}},
			"SELECT\n\tf.id, f.uuid, f.last_review, f.front, f.back, f.created_at, f.ext_data\nFROM\n\t" + constant.DATABASE_TABLE_FLASHCARDS + " f\n" +
				"WHERE\n" +
				"\t(\n" +
				"\t\tf.uuid = ?\n" +
				"\t);",
			[]any{"e3e9df30-bc8a-4458-af31-18fd437342fd"},
		},
		{
			"multiple id filters",
			args{parser.SearchFilters{Size: 3, IDs: []parser.Filter{
				{
					Type:      parser.ID,
					Key:       "",
					Value:     "1",
					IsExclude: false,
					Low:       1,
					High:      1,
				},
				{
					Type:      parser.ID,
					Key:       "",
					Value:     "17",
					IsExclude: false,
					Low:       17,
					High:      17,
				},
				{
					Type:      parser.ID,
					Key:       "",
					Value:     "10000",
					IsExclude: false,
					Low:       10000,
					High:      10000,
				},
			}}},
			"SELECT\n\tf.id, f.uuid, f.last_review, f.front, f.back, f.created_at, f.ext_data\nFROM\n\t" + constant.DATABASE_TABLE_FLASHCARDS + " f\n" +
				"WHERE\n" +
				"\t(\n" +
				"\t\tf.id = ?\n" +
				"\t\tOR f.id = ?\n" +
				"\t\tOR f.id = ?\n" +
				"\t);",
			[]any{1, 17, 10000},
		},
		{
			"range filter",
			args{parser.SearchFilters{Size: 1, Ranges: []parser.Filter{
				{
					Type:      parser.RANGE,
					Key:       "",
					Value:     "1-10",
					IsExclude: false,
					Low:       1,
					High:      10,
				},
			}}},
			"SELECT\n\tf.id, f.uuid, f.last_review, f.front, f.back, f.created_at, f.ext_data\nFROM\n\t" + constant.DATABASE_TABLE_FLASHCARDS + " f\n" +
				"WHERE\n" +
				"\t(\n" +
				"\t\tf.id BETWEEN ? AND ?\n" +
				"\t);",
			[]any{1, 10},
		},
		{
			"multiple range filters",
			args{parser.SearchFilters{Size: 2, Ranges: []parser.Filter{
				{
					Type:      parser.RANGE,
					Key:       "",
					Value:     "1-10",
					IsExclude: false,
					Low:       1,
					High:      10,
				},
				{
					Type:      parser.RANGE,
					Key:       "",
					Value:     "5-1000",
					IsExclude: false,
					Low:       5,
					High:      1000,
				},
			}}},
			"SELECT\n\tf.id, f.uuid, f.last_review, f.front, f.back, f.created_at, f.ext_data\nFROM\n\t" + constant.DATABASE_TABLE_FLASHCARDS + " f\n" +
				"WHERE\n" +
				"\t(\n" +
				"\t\tf.id BETWEEN ? AND ?\n" +
				"\t\tOR f.id BETWEEN ? AND ?\n" +
				"\t);",
			[]any{1, 10, 5, 1000},
		},
		{
			"group filter",
			args{parser.SearchFilters{Size: 1, Groups: []parser.Filter{
				{
					Type:      parser.GROUP,
					Key:       "group",
					Value:     "foo",
					IsExclude: false,
					Low:       -1,
					High:      -1,
				},
			}}},
			"SELECT\n\tf.id, f.uuid, f.last_review, f.front, f.back, f.created_at, f.ext_data\nFROM\n\t" + constant.DATABASE_TABLE_FLASHCARDS + " f\n" +
				"WHERE\n" +
				"\t(\n" +
				"\t\tEXISTS (SELECT 1 FROM " + constant.DATABASE_TABLE_GROUPS + " g WHERE f.id = g.flashcard_id AND g.name = ?)\n" +
				"\t);",
			[]any{"foo"},
		},
		{
			"multiple group filters",
			args{parser.SearchFilters{Size: 3, Groups: []parser.Filter{
				{
					Type:      parser.GROUP,
					Key:       "group",
					Value:     "foo",
					IsExclude: false,
					Low:       -1,
					High:      -1,
				},
				{
					Type:      parser.GROUP,
					Key:       "group",
					Value:     "bar",
					IsExclude: false,
					Low:       -1,
					High:      -1,
				},
				{
					Type:      parser.GROUP,
					Key:       "group",
					Value:     "baz",
					IsExclude: false,
					Low:       -1,
					High:      -1,
				},
			}}},
			"SELECT\n\tf.id, f.uuid, f.last_review, f.front, f.back, f.created_at, f.ext_data\nFROM\n\t" + constant.DATABASE_TABLE_FLASHCARDS + " f\n" +
				"WHERE\n" +
				"\t(\n" +
				"\t\tEXISTS (SELECT 1 FROM " + constant.DATABASE_TABLE_GROUPS + " g WHERE f.id = g.flashcard_id AND g.name = ?)\n" +
				"\t\tOR EXISTS (SELECT 1 FROM " + constant.DATABASE_TABLE_GROUPS + " g WHERE f.id = g.flashcard_id AND g.name = ?)\n" +
				"\t\tOR EXISTS (SELECT 1 FROM " + constant.DATABASE_TABLE_GROUPS + " g WHERE f.id = g.flashcard_id AND g.name = ?)\n" +
				"\t);",
			[]any{"foo", "bar", "baz"},
		},
		{
			"+ tag filter",
			args{parser.SearchFilters{Size: 1, Tags: []parser.Filter{
				{
					Type:      parser.TAG,
					Key:       "+",
					Value:     "foo",
					IsExclude: false,
					Low:       -1,
					High:      -1,
				},
			}}},
			"SELECT\n\tf.id, f.uuid, f.last_review, f.front, f.back, f.created_at, f.ext_data\nFROM\n\t" + constant.DATABASE_TABLE_FLASHCARDS + " f\n" +
				"WHERE\n" +
				"\t(\n" +
				"\t\tEXISTS (SELECT 1 FROM " + constant.DATABASE_TABLE_TAGS + " t WHERE f.id = t.flashcard_id AND t.name = ?)\n" +
				"\t);",
			[]any{"foo"},
		},
		{
			"- tag filter",
			args{parser.SearchFilters{Size: 1, Tags: []parser.Filter{
				{
					Type:      parser.TAG,
					Key:       "-",
					Value:     "foo",
					IsExclude: true,
					Low:       -1,
					High:      -1,
				},
			}}},
			"SELECT\n\tf.id, f.uuid, f.last_review, f.front, f.back, f.created_at, f.ext_data\nFROM\n\t" + constant.DATABASE_TABLE_FLASHCARDS + " f\n" +
				"WHERE\n" +
				"\t(\n" +
				"\t\tNOT EXISTS (SELECT 1 FROM " + constant.DATABASE_TABLE_TAGS + " t WHERE f.id = t.flashcard_id AND t.name = ?)\n" +
				"\t);",
			[]any{"foo"},
		},
		{
			"mixed tag filters",
			args{parser.SearchFilters{Size: 3, Tags: []parser.Filter{
				{
					Type:      parser.TAG,
					Key:       "+",
					Value:     "foo",
					IsExclude: false,
					Low:       -1,
					High:      -1,
				},
				{
					Type:      parser.TAG,
					Key:       "-",
					Value:     "bar",
					IsExclude: true,
					Low:       -1,
					High:      -1,
				},
				{
					Type:      parser.TAG,
					Key:       "+",
					Value:     "baz",
					IsExclude: false,
					Low:       -1,
					High:      -1,
				},
			}}},
			"SELECT\n\tf.id, f.uuid, f.last_review, f.front, f.back, f.created_at, f.ext_data\nFROM\n\t" + constant.DATABASE_TABLE_FLASHCARDS + " f\n" +
				"WHERE\n" +
				"\t(\n" +
				"\t\tEXISTS (SELECT 1 FROM " + constant.DATABASE_TABLE_TAGS + " t WHERE f.id = t.flashcard_id AND t.name = ?)\n" +
				"\t\tAND NOT EXISTS (SELECT 1 FROM " + constant.DATABASE_TABLE_TAGS + " t WHERE f.id = t.flashcard_id AND t.name = ?)\n" +
				"\t\tAND EXISTS (SELECT 1 FROM " + constant.DATABASE_TABLE_TAGS + " t WHERE f.id = t.flashcard_id AND t.name = ?)\n" +
				"\t);",
			[]any{"foo", "bar", "baz"},
		},
		{
			"mixed un/mandated filters",
			args{parser.SearchFilters{
				Size: 2,
				Groups: []parser.Filter{
					{
						Type:      parser.GROUP,
						Key:       "group",
						Value:     "foo",
						IsExclude: false,
						Low:       -1,
						High:      -1,
					},
				},
				Tags: []parser.Filter{
					{
						Type:      parser.TAG,
						Key:       "+",
						Value:     "bar",
						IsExclude: false,
						Low:       -1,
						High:      -1,
					},
				},
			}},
			"SELECT\n\tf.id, f.uuid, f.last_review, f.front, f.back, f.created_at, f.ext_data\nFROM\n\t" + constant.DATABASE_TABLE_FLASHCARDS + " f\n" +
				"WHERE\n" +
				"\t(\n" +
				"\t\tEXISTS (SELECT 1 FROM " + constant.DATABASE_TABLE_GROUPS + " g WHERE f.id = g.flashcard_id AND g.name = ?)\n" +
				"\t)\n" +
				"\tAND (\n" +
				"\t\tEXISTS (SELECT 1 FROM " + constant.DATABASE_TABLE_TAGS + " t WHERE f.id = t.flashcard_id AND t.name = ?)\n" +
				"\t);",
			[]any{"foo", "bar"},
		},
		{
			"all mixed filters",
			args{
				parser.SearchFilters{
					Size: 2,
					IDs: []parser.Filter{
						{
							Type:      parser.ID,
							Key:       "",
							Value:     "1",
							IsExclude: false,
							Low:       1,
							High:      1,
						},
						{
							Type:      parser.ID,
							Key:       "",
							Value:     "1337",
							IsExclude: false,
							Low:       1337,
							High:      1337,
						},
					},
					Ranges: []parser.Filter{
						{
							Type:      parser.RANGE,
							Key:       "",
							Value:     "2009-2330",
							IsExclude: false,
							Low:       2009,
							High:      2330,
						},
					},
					UUIDs: []parser.Filter{
						{
							Type:      parser.UUID,
							Key:       "",
							Value:     "e3e9df30-bc8a-4458-af31-18fd437342fd",
							IsExclude: false,
							Low:       -1,
							High:      -1,
						},
					},
					Groups: []parser.Filter{
						{
							Type:      parser.GROUP,
							Key:       "group",
							Value:     "foo",
							IsExclude: false,
							Low:       -1,
							High:      -1,
						},
						{
							Type:      parser.GROUP,
							Key:       "group",
							Value:     "bar",
							IsExclude: false,
							Low:       -1,
							High:      -1,
						},
					},
					Tags: []parser.Filter{
						{
							Type:      parser.TAG,
							Key:       "+",
							Value:     "qux",
							IsExclude: false,
							Low:       -1,
							High:      -1,
						},
						{
							Type:      parser.TAG,
							Key:       "-",
							Value:     "quux",
							IsExclude: true,
							Low:       -1,
							High:      -1,
						},
						{
							Type:      parser.TAG,
							Key:       "+",
							Value:     "corge",
							IsExclude: false,
							Low:       -1,
							High:      -1,
						},
					},
					// this doesn't get included in the search at the moment
					// may need to alter test want and add more tests to solo this out
					// if this behavior changes
					Customs: []parser.Filter{
						{
							Type:      parser.CUSTOM,
							Key:       "bing",
							Value:     "bong",
							IsExclude: false,
							Low:       -1,
							High:      -1,
						},
					},
				},
			},
			"SELECT\n\tf.id, f.uuid, f.last_review, f.front, f.back, f.created_at, f.ext_data\nFROM\n\t" + constant.DATABASE_TABLE_FLASHCARDS + " f\n" +
				"WHERE\n" +
				"\t(\n" +
				"\t\tf.id = ?\n" +
				"\t\tOR f.id = ?\n" +
				"\t\tOR f.id BETWEEN ? AND ?\n" +
				"\t\tOR f.uuid = ?\n" +
				"\t\tOR EXISTS (SELECT 1 FROM " + constant.DATABASE_TABLE_GROUPS + " g WHERE f.id = g.flashcard_id AND g.name = ?)\n" +
				"\t\tOR EXISTS (SELECT 1 FROM " + constant.DATABASE_TABLE_GROUPS + " g WHERE f.id = g.flashcard_id AND g.name = ?)\n" +
				"\t)\n" +
				"\tAND (\n" +
				"\t\tEXISTS (SELECT 1 FROM " + constant.DATABASE_TABLE_TAGS + " t WHERE f.id = t.flashcard_id AND t.name = ?)\n" +
				"\t\tAND NOT EXISTS (SELECT 1 FROM " + constant.DATABASE_TABLE_TAGS + " t WHERE f.id = t.flashcard_id AND t.name = ?)\n" +
				"\t\tAND EXISTS (SELECT 1 FROM " + constant.DATABASE_TABLE_TAGS + " t WHERE f.id = t.flashcard_id AND t.name = ?)\n" +
				"\t);",
			// another note on CUSTOM filters, they are also not included in the args
			[]any{1, 1337, 2009, 2330, "e3e9df30-bc8a-4458-af31-18fd437342fd", "foo", "bar", "qux", "quux", "corge"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := buildFlashcardSelectQuery(tt.args.filters)
			if got != tt.want {
				t.Errorf("buildFlashcardSelectQuery() got = \n```\n%v\n```\nwant =\n```\n%v\n```\n\nDIFF: %v", got, tt.want, cmp.Diff(got, tt.want))
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("buildFlashcardSelectQuery() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
