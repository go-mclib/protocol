package textformat

import (
	"slices"
	"strings"

	ns "github.com/go-mclib/protocol/javaproto/wire"
)

// MiniMessage returns tc in Adventure MiniMessage format. Translatable
// components become <lang:key:args> and keybinds <key:name>.
func MiniMessage(tc ns.TextComponent) string {
	var b strings.Builder
	writeMiniMessage(&b, &tc)
	return b.String()
}

func writeMiniMessage(b *strings.Builder, tc *ns.TextComponent) {
	var tags []string

	if tc.Color != "" {
		tags = append(tags, tc.Color)
	}
	if tc.Bold != nil && *tc.Bold {
		tags = append(tags, "bold")
	}
	if tc.Italic != nil && *tc.Italic {
		tags = append(tags, "italic")
	}
	if tc.Underlined != nil && *tc.Underlined {
		tags = append(tags, "underlined")
	}
	if tc.Strikethrough != nil && *tc.Strikethrough {
		tags = append(tags, "strikethrough")
	}
	if tc.Obfuscated != nil && *tc.Obfuscated {
		tags = append(tags, "obfuscated")
	}

	for _, tag := range tags {
		b.WriteByte('<')
		b.WriteString(tag)
		b.WriteByte('>')
	}

	if tc.Translate != "" {
		b.WriteString("<lang:")
		b.WriteString(tc.Translate)
		for i := range tc.With {
			b.WriteByte(':')
			writeMiniMessage(b, &tc.With[i])
		}
		b.WriteByte('>')
	} else if tc.Object != nil {
		if tc.Object.Fallback != nil {
			writeMiniMessage(b, tc.Object.Fallback)
		} else {
			// an object without fallback renders as its plain default fallback
			b.WriteString(ns.TextComponent{Object: tc.Object}.String())
		}
	} else if tc.Keybind != "" {
		b.WriteString("<key:")
		b.WriteString(tc.Keybind)
		b.WriteByte('>')
	} else {
		b.WriteString(tc.Text)
		if tc.Score != nil {
			b.WriteString(tc.Score.Name)
		}
		b.WriteString(tc.Selector)
	}

	for i := range tc.Extra {
		writeMiniMessage(b, &tc.Extra[i])
	}

	// close tags in reverse
	for _, tag := range slices.Backward(tags) {
		b.WriteString("</")
		b.WriteString(tag)
		b.WriteByte('>')
	}
}
