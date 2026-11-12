package textformat_test

import (
	"slices"
	"testing"

	"github.com/go-mclib/protocol/javaproto/textformat"
	ns "github.com/go-mclib/protocol/javaproto/wire"
)

func TestParseFormattedUnformatted(t *testing.T) {
	tc := textformat.ParseFormatted("hello world")
	if tc.Text != "hello world" {
		t.Errorf("expected plain text, got %+v", tc)
	}
}

func TestParseFormattedLegacyColor(t *testing.T) {
	tc := textformat.ParseFormatted("&6gold text")
	if tc.Color != "gold" || tc.Text != "gold text" {
		t.Errorf("expected gold 'gold text', got color=%q text=%q extras=%d", tc.Color, tc.Text, len(tc.Extra))
	}
}

func TestParseFormattedLegacyMultipleColors(t *testing.T) {
	tc := textformat.ParseFormatted("&ared &bblue")
	if len(tc.Extra) != 2 {
		t.Fatalf("expected 2 extras, got %d", len(tc.Extra))
	}
	if tc.Extra[0].Color != "green" || tc.Extra[0].Text != "red " {
		t.Errorf("segment 0: %+v", tc.Extra[0])
	}
	if tc.Extra[1].Color != "aqua" || tc.Extra[1].Text != "blue" {
		t.Errorf("segment 1: %+v", tc.Extra[1])
	}
}

func TestParseFormattedLegacyBold(t *testing.T) {
	tc := textformat.ParseFormatted("&6&lbold gold")
	found := slices.ContainsFunc(flatExtras(tc), func(e ns.TextComponent) bool {
		return e.Text == "bold gold" && e.Bold != nil && *e.Bold && e.Color == "gold"
	})
	if !found {
		t.Errorf("expected bold gold text, got %+v", tc)
	}
}

func TestParseFormattedSectionSign(t *testing.T) {
	tc := textformat.ParseFormatted("§cred text")
	if tc.Color != "red" || tc.Text != "red text" {
		t.Errorf("expected red 'red text', got color=%q text=%q", tc.Color, tc.Text)
	}
}

func TestParseFormattedMiniMessageColor(t *testing.T) {
	tc := textformat.ParseFormatted("<red>hello</red>")
	found := slices.ContainsFunc(flatExtras(tc), func(e ns.TextComponent) bool { return e.Color == "red" })
	if !found {
		t.Errorf("expected red tag, got %+v", tc)
	}
}

func TestParseFormattedMiniMessageNested(t *testing.T) {
	tc := textformat.ParseFormatted("<bold><red>hello</red></bold>")
	// should have bold with red child
	if tc.Bold == nil || !*tc.Bold {
		t.Errorf("expected bold root, got %+v", tc)
	}
	plain := tc.String()
	if plain != "hello" {
		t.Errorf("expected 'hello' plain text, got %q", plain)
	}
}

func TestParseFormattedMixed(t *testing.T) {
	tc := textformat.ParseFormatted("&6gold <red>red</red> &baqua")
	plain := tc.String()
	if plain != "gold red aqua" {
		t.Errorf("expected 'gold red aqua', got %q", plain)
	}
}

func TestParseFormattedHexColor(t *testing.T) {
	tc := textformat.ParseFormatted("<#ff5555>custom</  #ff5555>")
	// the malformed closing tag is not found, so the color applies to the rest
	found := slices.ContainsFunc(flatExtras(tc), func(e ns.TextComponent) bool { return e.Color == "#ff5555" })
	if !found {
		t.Errorf("expected #ff5555 color, got %+v", tc)
	}
}

func TestParseFormattedReset(t *testing.T) {
	tc := textformat.ParseFormatted("&6gold&r plain")
	if len(tc.Extra) < 2 {
		t.Fatalf("expected at least 2 extras, got %d", len(tc.Extra))
	}
	last := tc.Extra[len(tc.Extra)-1]
	if last.Color != "" || last.Text != " plain" {
		t.Errorf("after reset: %+v", last)
	}
}

func TestParseFormattedNonASCII(t *testing.T) {
	for in, want := range map[string]string{
		"&a» é":        "» é",
		"&cПривет &l★": "Привет ★",
		"§6金 &x& ok":   "金 &x& ok",
	} {
		if got := textformat.ParseFormatted(in).String(); got != want {
			t.Errorf("ParseFormatted(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestParseFormattedStylesAreIndependent(t *testing.T) {
	first := textformat.ParseFormatted("&lone")
	*first.Bold = false
	if second := textformat.ParseFormatted("&ltwo"); second.Bold == nil || !*second.Bold {
		t.Fatalf("mutating one result changed another: %+v", second)
	}
}

// flatExtras collects all components in the tree (recursive).
func flatExtras(tc ns.TextComponent) []ns.TextComponent {
	var result []ns.TextComponent
	result = append(result, tc)
	for _, e := range tc.Extra {
		result = append(result, flatExtras(e)...)
	}
	return result
}
