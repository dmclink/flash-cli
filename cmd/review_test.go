package cmd

import (
	"reflect"
	"testing"

	_ "modernc.org/sqlite"
)

func Test_hasModPrefix(t *testing.T) {
	type args struct {
		s    string
		name string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"has : prefix", args{"foo:bar", "foo"}, true},
		{"has = prefix", args{"foo=bar", "foo"}, true},
		{"equals with delim", args{"foo=", "foo"}, true},
		{"equals without delim", args{"foo", "foo"}, false},
		{"has prefix but incorrect delim", args{"foo-bar", "foo"}, false},
		{"incorrect prefix", args{"bar=foo", "foo"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasModPrefix(tt.args.s, tt.args.name); got != tt.want {
				t.Errorf("hasModPrefix() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_findPrefixIdx(t *testing.T) {
	type args struct {
		mods   []string
		prefix string
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{"valid1", args{[]string{"foo:", "bar:", "baz:", "qux:"}, "foo"}, 0},
		{"valid2", args{[]string{"foo:qux", "bar:qux", "baz:qux", "qux:foo"}, "qux"}, 3},
		{"valid3", args{[]string{"foo:", "bar:", "baz:", "qux:"}, "baz"}, 2},
		{"valid4", args{[]string{"foo=", "bar=", "baz=", "qux="}, "baz"}, 2},
		{"multiple entries", args{[]string{"foo:", "bar:", "bar:", "qux:"}, "bar"}, 1},
		{"doesn't exist", args{[]string{"foo:", "bar:", "baz:", "qux:"}, "quux"}, -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := findPrefixIdx(tt.args.mods, tt.args.prefix); got != tt.want {
				t.Errorf("findPrefixIdx() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_stripModPrefix(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{": delim", args{"foo:bar"}, "bar"},
		{"= delim", args{"foo=bar"}, "bar"},
		{"no delim", args{"foobar"}, "foobar"},
		{"no value", args{"foo:"}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := stripModPrefix(tt.args.s); got != tt.want {
				t.Errorf("stripModPrefix() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_removeFromMods(t *testing.T) {
	type args struct {
		mods    []string
		modName string
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 []string
	}{
		{"found mode1", args{[]string{"mode=foo"}, "mode"}, "foo", []string{}},
		{"found mode2", args{[]string{"renderer=default", "mode:foo"}, "mode"}, "foo", []string{"renderer=default"}},
		{"no mode", args{[]string{"renderer=default", "mod=foo"}, "mode"}, "", []string{"renderer=default", "mod=foo"}},
		{"empty mods", args{[]string{}, "mode"}, "", []string{}},
		{"empty mode", args{[]string{"mode:"}, "mode"}, "", []string{}},
		{"multiple mode", args{[]string{"mode:foo", "mode:bar", "renderer:baz", "mode:qux"}, "mode"}, "qux", []string{"renderer:baz"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := removeFromMods(tt.args.mods, tt.args.modName)
			if got != tt.want {
				t.Errorf("removeFromMods() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("removeFromMods() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_removeRenderer(t *testing.T) {
	type args struct {
		mods []string
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 []string
	}{
		{"no renderer", args{[]string{"mode=foo"}}, "", []string{"mode=foo"}},
		{"found renderer", args{[]string{"renderer=default", "mode:foo"}}, "default", []string{"mode:foo"}},
		{"multiple renderer", args{[]string{"renderer=default", "mode:foo", "renderer:bar"}}, "bar", []string{"mode:foo"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := removeRenderer(tt.args.mods)
			if got != tt.want {
				t.Errorf("removeRenderer() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("removeRenderer() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_removeMode(t *testing.T) {
	type args struct {
		mods []string
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 []string
	}{
		{"no mode", args{[]string{"renderer=foo"}}, "", []string{"renderer=foo"}},
		{"found mode", args{[]string{"mode=default", "renderer:foo"}}, "default", []string{"renderer:foo"}},
		{"multiple mode", args{[]string{"mode=default", "renderer:foo", "mode:bar"}}, "bar", []string{"renderer:foo"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := removeMode(tt.args.mods)
			if got != tt.want {
				t.Errorf("removeMode() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("removeMode() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
