package parser

import (
	"reflect"
	"slices"
	"sort"
	"testing"
)

func TestParseFilters(t *testing.T) {
	type args struct {
		args ParsedArgs
	}
	tests := []struct {
		name string
		args args
		want []Filter
	}{
		{
			"single",
			args{ParsedArgs{Command: "add", Filters: []string{"group:foo"}, Mods: []string{"new card::back"}, OriginalInput: "group:foo add new card::back"}},
			[]Filter{{"group:foo", GROUP, nil}},
		},
		{
			"compound ids",
			args{ParsedArgs{Command: "review", Filters: []string{"1,2,8-10", "20"}, Mods: []string{}, OriginalInput: "1,2,8-10 20 review"}},
			[]Filter{{"1", ID, nil}, {"2", ID, nil}, {"8-10", RANGE, &Range{8, 10}}, {"20", ID, nil}},
		},
		{
			"compound groups",
			args{ParsedArgs{Command: "review", Filters: []string{"group:foo,bar"}, Mods: []string{}, OriginalInput: "group:foo,bar review"}},
			[]Filter{{"group:foo", GROUP, nil}, {"group:bar", GROUP, nil}},
		},
		{
			"compound group alias",
			args{ParsedArgs{Command: "review", Filters: []string{"grp:foo,bar"}, Mods: []string{}, OriginalInput: "group:foo,bar review"}},
			[]Filter{{"group:foo", GROUP, nil}, {"group:bar", GROUP, nil}},
		},
		{
			"compound custom",
			args{ParsedArgs{Command: "review", Filters: []string{"baz:foo,bar"}, Mods: []string{}, OriginalInput: "group:foo,bar review"}},
			[]Filter{{"baz:foo", CUSTOM, nil}, {"baz:bar", CUSTOM, nil}},
		},
		{
			"UUID starts with digit",
			args{ParsedArgs{Command: "review", Filters: []string{"0fb80f43-cb89-4d21-a5a1-7ef2995e7306"}, Mods: []string{}, OriginalInput: "0fb80f43-cb89-4d21-a5a1-7ef2995e7306 review"}},
			[]Filter{{"0fb80f43-cb89-4d21-a5a1-7ef2995e7306", UUID, nil}},
		},
		{
			"UUID starts with alpha",
			args{ParsedArgs{Command: "review", Filters: []string{"e3e9df30-bc8a-4458-af31-18fd437342fd"}, Mods: []string{}, OriginalInput: "e3e9df30-bc8a-4458-af31-18fd437342fd review"}},
			[]Filter{{"e3e9df30-bc8a-4458-af31-18fd437342fd", UUID, nil}},
		},
		{
			"compound tags",
			args{ParsedArgs{Command: "review", Filters: []string{"+foo,bar"}, Mods: []string{}, OriginalInput: "+foo,bar review"}},
			[]Filter{{"+foo", TAG, nil}, {"+bar", TAG, nil}},
		},
		{
			"- tag",
			args{ParsedArgs{Command: "review", Filters: []string{"-foo"}, Mods: []string{}, OriginalInput: "-foo review"}},
			[]Filter{{"-foo", TAG, nil}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseFilters(tt.args.args)
			// need sorting since implementation uses maps and order isnt preserved
			sort.Slice(got, func(i, j int) bool {
				return got[i].f < got[j].f
			})
			sort.Slice(tt.want, func(i, j int) bool {
				return tt.want[i].f < tt.want[j].f
			})
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseFilters() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilter_String(t *testing.T) {
	type fields struct {
		f     string
		Type  FilterType
		Range *Range
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"empty string", fields{"", CUSTOM, nil}, ""},
		{"valid filter", fields{"group:foo", GROUP, nil}, "group:foo"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := Filter{
				f:    tt.fields.f,
				Type: tt.fields.Type,
			}
			if got := f.String(); got != tt.want {
				t.Errorf("Filter.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewRange(t *testing.T) {
	type args struct {
		f RawFilter
	}
	tests := []struct {
		name string
		args args
		want Range
	}{
		{"same number", args{"1-1"}, Range{1, 1}},
		{"same double digit number", args{"10-10"}, Range{10, 10}},
		{"regular range", args{"5-10"}, Range{5, 10}},
		{"larger number first", args{"10-5"}, Range{5, 10}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewRange(tt.args.f); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewRange() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_baseFilter_String(t *testing.T) {
	type fields struct {
		f RawFilter
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"empty string", fields{""}, ""},
		{"valid filter", fields{"group:foo"}, "group:foo"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := baseFilter{
				f: tt.fields.f,
			}
			if got := b.String(); got != tt.want {
				t.Errorf("baseFilter.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRawFilter_String(t *testing.T) {
	tests := []struct {
		name string
		rf   RawFilter
		want string
	}{
		{"empty string", "", ""},
		{"valid filter", "group:foo", "group:foo"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.rf.String(); got != tt.want {
				t.Errorf("RawFilter.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_splitAtCommas(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{"empty string", args{""}, []string{""}},
		{"ids", args{"1,5,10-100"}, []string{"1", "5", "10-100"}},
		{"groups", args{"foo,Bar,BAZ"}, []string{"foo", "Bar", "BAZ"}},
		{"duplicates", args{"foo,Bar,foo"}, []string{"foo", "Bar"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// need to sort output for test since order is not maintained
			got := splitAtCommas(tt.args.s)
			slices.Sort(got)
			slices.Sort(tt.want)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("splitAtCommas() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_toRawFiltersWithPrefix(t *testing.T) {
	type args struct {
		filters []string
		prefix  string
	}
	tests := []struct {
		name string
		args args
		want []RawFilter
	}{
		{"ids", args{[]string{"1", "3-5"}, ""}, []RawFilter{"1", "3-5"}},
		{"groups", args{[]string{"foo", "bar"}, "group:"}, []RawFilter{"group:foo", "group:bar"}},
		{"tags", args{[]string{"foo", "bar"}, "+"}, []RawFilter{"+foo", "+bar"}},
		{"custom", args{[]string{"foo", "bar"}, "baz:"}, []RawFilter{"baz:foo", "baz:bar"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := toRawFiltersWithPrefix(tt.args.filters, tt.args.prefix); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("toRawFiltersWithPrefix() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_compoundGroupFilter_split(t *testing.T) {
	type fields struct {
		baseFilter baseFilter
	}
	tests := []struct {
		name   string
		fields fields
		want   []RawFilter
	}{
		{"group prefix", fields{baseFilter{"group:foo,bar"}}, []RawFilter{"group:foo", "group:bar"}},
		{"grp prefix", fields{baseFilter{"grp:foo,bar"}}, []RawFilter{"group:foo", "group:bar"}},
		{"groups prefix", fields{baseFilter{"groups:foo,bar"}}, []RawFilter{"group:foo", "group:bar"}},
		{"project prefix", fields{baseFilter{"project:foo,bar"}}, []RawFilter{"group:foo", "group:bar"}},
		{"proj prefix", fields{baseFilter{"proj:foo,bar"}}, []RawFilter{"group:foo", "group:bar"}},
		{"prj prefix", fields{baseFilter{"prj:foo,bar"}}, []RawFilter{"group:foo", "group:bar"}},
		{"duplicate with group prefix", fields{baseFilter{"group:foo,bar,foo"}}, []RawFilter{"group:foo", "group:bar"}},
		{"duplicate with proj prefix", fields{baseFilter{"proj:foo,bar,foo"}}, []RawFilter{"group:foo", "group:bar"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := compoundGroupFilter{
				baseFilter: tt.fields.baseFilter,
			}
			got := f.split()
			slices.Sort(got)
			slices.Sort(tt.want)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("compoundGroupFilter.split() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_compoundIDFilter_split(t *testing.T) {
	type fields struct {
		baseFilter baseFilter
	}
	tests := []struct {
		name   string
		fields fields
		want   []RawFilter
	}{
		{"no ranges", fields{baseFilter{"1,5,10"}}, []RawFilter{"1", "5", "10"}},
		{"ranges", fields{baseFilter{"1,5,10-20,30-50"}}, []RawFilter{"1", "5", "10-20", "30-50"}},
		{"duplicate ids and ranges", fields{baseFilter{"1,5,5,10-20,10-20,30-50"}}, []RawFilter{"1", "5", "10-20", "30-50"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := compoundIDFilter{
				baseFilter: tt.fields.baseFilter,
			}

			got := f.split()
			slices.Sort(got)
			slices.Sort(tt.want)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("compoundIDFilter.split() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_compoundTagFilter_split(t *testing.T) {
	type fields struct {
		baseFilter baseFilter
	}
	tests := []struct {
		name   string
		fields fields
		want   []RawFilter
	}{
		{"+ tags", fields{baseFilter{"+foo,bar,baz"}}, []RawFilter{"+foo", "+bar", "+baz"}},
		{"- tags", fields{baseFilter{"-foo,bar,baz"}}, []RawFilter{"-foo", "-bar", "-baz"}},
		{"duplicate tags", fields{baseFilter{"-foo,foo,foo"}}, []RawFilter{"-foo"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := compoundTagFilter{
				baseFilter: tt.fields.baseFilter,
			}

			got := f.split()
			slices.Sort(got)
			slices.Sort(tt.want)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("compoundTagFilter.split() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_compoundCustomFilter_split(t *testing.T) {
	type fields struct {
		baseFilter baseFilter
	}
	tests := []struct {
		name   string
		fields fields
		want   []RawFilter
	}{
		{"custom filter", fields{baseFilter{"foo:bar,baz"}}, []RawFilter{"foo:bar", "foo:baz"}},
		{"empty custom filter", fields{baseFilter{":,,,"}}, []RawFilter{":"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := compoundCustomFilter{
				baseFilter: tt.fields.baseFilter,
			}

			got := f.split()
			slices.Sort(got)
			slices.Sort(tt.want)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("compoundCustomFilter.split() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRawFilter_getType(t *testing.T) {
	tests := []struct {
		name   string
		filter RawFilter
		want   FilterType
	}{
		{"+ tag", "+foo", TAG},
		{"- tag", "-foo", TAG},
		{"- group tag", "-foo,foo", TAG},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.filter.getType(); got != tt.want {
				t.Errorf("RawFilter.getType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRawFilter_isCompound(t *testing.T) {
	tests := []struct {
		name   string
		filter RawFilter
		want   bool
	}{
		{"compound + tag", "+foo,bar", true},
		{"compound - tag", "-foo,bar", true},
		{"compound group", "group:foo,bar", true},
		{"compound custom", "baz:foo,bar", true},
		{"compound id", "1,4,6-10", true},
		{"single id", "14", false},
		{"single range", "14-20", false},
		{"single group", "group:foo", false},
		{"single tag", "+foo", false},
		{"single custom", "custom:foo", false},
		{"single uuid starts with digit", "0fb80f43-cb89-4d21-a5a1-7ef2995e7306", false},
		{"single uuid starts with alpha", "e3e9df30-bc8a-4458-af31-18fd437342fd", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.filter.isCompound(); got != tt.want {
				t.Errorf("RawFilter.isCompound() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRawFilter_isGroupType(t *testing.T) {
	tests := []struct {
		name   string
		filter RawFilter
		want   bool
	}{
		{"group prefix", "group:foo", true},
		{"grp prefix", "grp:foo", true},
		{"groups prefix", "groups:foo", true},
		{"project prefix", "project:foo", true},
		{"proj prefix", "proj:foo", true},
		{"prj prefix", "prj:foo", true},
		{"group prefix compound", "group:foo,bar", true},
		{"id", "1", false},
		{"range", "5-10", false},
		{"custom", "foo:bar", false},
		{"+ tag", "+foo", false},
		{"- tag", "-foo", false},
		{"tag containing :", "+foo:bar", false},
		{"uuid starts with digit", "0fb80f43-cb89-4d21-a5a1-7ef2995e7306", false},
		{"uuid starts with alpha", "e3e9df30-bc8a-4458-af31-18fd437342fd", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.filter.isGroupType(); got != tt.want {
				t.Errorf("RawFilter.isGroupType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRawFilter_isTagType(t *testing.T) {
	tests := []struct {
		name   string
		filter RawFilter
		want   bool
	}{
		{"+ tag", "+foo", true},
		{"- tag", "-foo", true},
		{"tag containing :", "+foo:bar", true},
		{"tag compound", "+foo,bar", true},
		{"group prefix", "group:foo", false},
		{"grp prefix", "grp:foo", false},
		{"groups prefix", "groups:foo", false},
		{"project prefix", "project:foo", false},
		{"proj prefix", "proj:foo", false},
		{"prj prefix", "prj:foo", false},
		{"group prefix compound", "group:foo,bar", false},
		{"id", "1", false},
		{"range", "5-10", false},
		{"custom", "foo:bar", false},
		{"uuid starts with digit", "0fb80f43-cb89-4d21-a5a1-7ef2995e7306", false},
		{"uuid starts with alpha", "e3e9df30-bc8a-4458-af31-18fd437342fd", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.filter.isTagType(); got != tt.want {
				t.Errorf("RawFilter.isTagType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRawFilter_isCustomDataType(t *testing.T) {
	tests := []struct {
		name   string
		filter RawFilter
		want   bool
	}{
		{"custom", "foo:bar", true},
		{"custom compound", "foo:bar", true},
		{"only :", ":", true},
		// the following true tests result to true because of the simplistic implementation of
		// isCustomDataType. it only checks for existence of a semicolon ':' character in the string
		// this is why it is imperative to rule out tags and groups before calling this function in production
		{"tag containing :", "+foo:bar", true},
		{"group prefix", "group:foo", true},
		{"grp prefix", "grp:foo", true},
		{"groups prefix", "groups:foo", true},
		{"project prefix", "project:foo", true},
		{"proj prefix", "proj:foo", true},
		{"prj prefix", "prj:foo", true},
		{"group prefix compound", "group:foo,bar", true},

		{"+ tag", "+foo", false},
		{"- tag", "-foo", false},
		{"tag compound", "+foo,bar", false},
		{"id", "1", false},
		{"range", "5-10", false},
		{"uuid starts with digit", "0fb80f43-cb89-4d21-a5a1-7ef2995e7306", false},
		{"uuid starts with alpha", "e3e9df30-bc8a-4458-af31-18fd437342fd", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.filter.isCustomDataType(); got != tt.want {
				t.Errorf("RawFilter.isCustomDataType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRawFilter_isUUIDType(t *testing.T) {
	tests := []struct {
		name   string
		filter RawFilter
		want   bool
	}{
		{"uuid starts with digit", "0fb80f43-cb89-4d21-a5a1-7ef2995e7306", true},
		{"uuid starts with alpha", "e3e9df30-bc8a-4458-af31-18fd437342fd", true},
		{"+ tag", "+foo", false},
		{"- tag", "-foo", false},
		{"tag containing :", "+foo:bar", false},
		{"tag compound", "+foo,bar", false},
		{"group prefix", "group:foo", false},
		{"grp prefix", "grp:foo", false},
		{"groups prefix", "groups:foo", false},
		{"project prefix", "project:foo", false},
		{"proj prefix", "proj:foo", false},
		{"prj prefix", "prj:foo", false},
		{"group prefix compound", "group:foo,bar", false},
		{"id", "1", false},
		{"range", "5-10", false},
		{"custom", "foo:bar", false},
		{"only :", ":", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.filter.isUUIDType(); got != tt.want {
				t.Errorf("RawFilter.isUUIDType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRawFilter_isRangeType(t *testing.T) {
	tests := []struct {
		name   string
		filter RawFilter
		want   bool
	}{
		{"range", "5-10", true},
		// note that a range is an ID type but an ID is not a range
		{"id", "1", false},
		{"compound ranges", "1-5,10-50", false}, // nor is a compound containing a range

		{"uuid starts with digit", "0fb80f43-cb89-4d21-a5a1-7ef2995e7306", false},
		{"uuid starts with alpha", "e3e9df30-bc8a-4458-af31-18fd437342fd", false},
		{"+ tag", "+foo", false},
		{"- tag", "-foo", false},
		{"tag containing :", "+foo:bar", false},
		{"tag compound", "+foo,bar", false},
		{"group prefix", "group:foo", false},
		{"grp prefix", "grp:foo", false},
		{"groups prefix", "groups:foo", false},
		{"project prefix", "project:foo", false},
		{"proj prefix", "proj:foo", false},
		{"prj prefix", "prj:foo", false},
		{"group prefix compound", "group:foo,bar", false},
		{"custom", "foo:bar", false},
		{"only :", ":", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.filter.isRangeType(); got != tt.want {
				t.Errorf("RawFilter.isRangeType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRawFilter_isIDType(t *testing.T) {
	tests := []struct {
		name   string
		filter RawFilter
		want   bool
	}{
		{"range", "5-10", true},
		// note that a range is an ID type but an ID is not a range
		{"id", "1", true},
		{"compound ranges", "1-5,10-50", true},

		// UUIDs starting with digit also resolve to true, important to test for valid UUID before this function
		{"uuid starts with digit", "0fb80f43-cb89-4d21-a5a1-7ef2995e7306", true},

		// another edge case which requires us to check for CUSTOM before ID
		{"custom that looks like an id", "1:2", true},

		{"uuid starts with alpha", "e3e9df30-bc8a-4458-af31-18fd437342fd", false},
		{"+ tag", "+foo", false},
		{"- tag", "-foo", false},
		{"tag containing :", "+foo:bar", false},
		{"tag compound", "+foo,bar", false},
		{"group prefix", "group:foo", false},
		{"grp prefix", "grp:foo", false},
		{"groups prefix", "groups:foo", false},
		{"project prefix", "project:foo", false},
		{"proj prefix", "proj:foo", false},
		{"prj prefix", "prj:foo", false},
		{"group prefix compound", "group:foo,bar", false},
		{"custom", "foo:bar", false},
		{"only :", ":", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.filter.isIDType(); got != tt.want {
				t.Errorf("RawFilter.isIDType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetGroups(t *testing.T) {
	type args struct {
		filters []Filter
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			"no filters",
			args{[]Filter{}},
			[]string{},
		},
		{
			"no groups",
			args{[]Filter{{"1", ID, nil}, {"4-9", RANGE, &Range{4, 9}}, {"foo:bar", CUSTOM, nil}}},
			[]string{},
		},
		{
			"one group",
			args{[]Filter{{"1", ID, nil}, {"4-9", RANGE, &Range{4, 9}}, {"group:baz", GROUP, nil}, {"foo:bar", CUSTOM, nil}}},
			[]string{"baz"},
		},
		{
			"many groups",
			args{[]Filter{{"group:foo", GROUP, nil}, {"group:baz", GROUP, nil}, {"group:bar", CUSTOM, nil}}},
			[]string{"foo", "baz", "bar"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetGroups(tt.args.filters); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetGroups() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetTags(t *testing.T) {
	type args struct {
		filters []Filter
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetTags(tt.args.filters); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetTags() = %v, want %v", got, tt.want)
			}
		})
	}
}
