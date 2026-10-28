package app

import (
	"reflect"
	"testing"

	tcell "github.com/gdamore/tcell/v2"
	"github.com/yairgd/termforge"
	"github.com/yairgd/termforge/commands"
)

func TestCompleteLastWord(t *testing.T) {
	cols := []string{"pe_ratio", "price", "sector", "date"}
	tests := []struct {
		name     string
		prefix   string
		emptyOK  bool
		maxWords int
		want     []string
	}{
		{"first word", "p", false, 0, []string{"pe_ratio", "price"}},
		{"empty filter lists all", "", false, 0, cols},
		{"after and", "pe_ratio < 10 and sec", false, 0, []string{"sector"}},
		{"inside parens", "(pe_ratio<5) or (pr", false, 0, []string{"(price"}},
		{"in quotes", "sector = 'pe", false, 0, nil},
		{"number", "pe_ratio > 1", false, 0, nil},
		{"empty after operator", "pe_ratio > ", false, 0, nil},
		{"plot second field", "price d", true, 2, []string{"date"}},
		{"plot second field empty", "price ", true, 2, cols},
		{"plot third word", "price date x", true, 2, nil},
	}
	for _, tc := range tests {
		got := completeLastWord(cols, tc.prefix, tc.emptyOK, tc.maxWords)
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("%s: completeLastWord(%q) = %v, want %v", tc.name, tc.prefix, got, tc.want)
		}
	}
}

// Tab on the real command line: the typed text must keep the earlier words.
func TestTabCompletesInsideFilterAndPlot(t *testing.T) {
	cols := []string{"pe_ratio", "price", "sector", "date"}
	reg := commands.NewCommandRegistry()
	reg.Root.
		LeafRestComplete("filter", func(...any) {}, func(p string, _ bool) []string { return completeLastWord(cols, p, false, 0) }).
		LeafRestComplete("plot", func(...any) {}, func(p string, _ bool) []string { return completeLastWord(cols, p, true, 2) })
	run := func(typed string) string {
		w := termforge.NewCmdWidget(reg)
		w.Activate()
		for _, r := range typed {
			w.HandleEvent(tcell.NewEventKey(tcell.KeyRune, r, 0))
		}
		w.HandleEvent(tcell.NewEventKey(tcell.KeyTAB, 0, 0))
		return w.Text()
	}
	for typed, want := range map[string]string{
		"filte":                       ":filter ",
		"filter pe_ratio < 10 and se": ":filter pe_ratio < 10 and sector",
		"filter (pe_ratio<5) or (pr":  ":filter (pe_ratio<5) or (price",
		"plot price d":                ":plot price date",
		"filter sector = 'pe":         ":filter sector = 'pe",
	} {
		if got := run(typed); got != want {
			t.Errorf("Tab after %q: got %q, want %q", typed, got, want)
		}
	}
}

func TestCommonPrefixAndLastChunk(t *testing.T) {
	if got := commonPrefix([]string{"pe_ratio", "PE_ttm", "price"}); got != "p" {
		t.Errorf("commonPrefix = %q, want p", got)
	}
	if got := commonPrefix([]string{"ev_ebit", "ev_ebitda"}); got != "ev_ebit" {
		t.Errorf("commonPrefix = %q, want ev_ebit", got)
	}
	for in, want := range map[string]string{":fil": "fil", ":filter pe < 1 and se": "se", ":filter ": "", ":": ""} {
		if got := lastChunk(in); got != want {
			t.Errorf("lastChunk(%q) = %q, want %q", in, got, want)
		}
	}
}
