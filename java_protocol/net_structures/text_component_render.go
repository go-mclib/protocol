package net_structures

import (
	"fmt"
	"strings"
)

// String returns the plain text content of the component and all children,
// with no formatting. Translate keys are shown as-is.
func (tc TextComponent) String() string {
	return tc.Render(nil)
}

// Render returns the plain text with translate keys resolved by fn (if non-nil).
func (tc TextComponent) Render(translate func(string) string) string {
	var b strings.Builder
	tc.writePlain(&b, translate)
	return b.String()
}

// MiniMessage returns the text in Adventure MiniMessage format.
// Translate keys are shown as-is.
func (tc TextComponent) MiniMessage() string {
	var b strings.Builder
	tc.writeMiniMessage(&b)
	return b.String()
}

// PlainText extracts the plain text content from a TextComponent tree,
// stripping all formatting.
func (tc *TextComponent) PlainText() string {
	var sb strings.Builder
	tc.plainTextInto(&sb)
	return sb.String()
}

// componentWriter writes a single component's full content (content + extras) to b.
type componentWriter func(tc *TextComponent, b *strings.Builder)

// writeContent writes the resolved content of this component (without extras).
// If translate is non-nil and the component has a translate key, the key is
// resolved and %s / %N$s placeholders are substituted with With args rendered
// via the provided writer. Otherwise the raw text/key is written.
func (tc *TextComponent) writeContent(b *strings.Builder, write componentWriter, translate func(string) string) {
	if tc.Translate != "" {
		if translate != nil {
			if pattern := translate(tc.Translate); pattern != "" {
				writeFormatted(b, pattern, tc.With, write)
				return
			}
		}
		// no translation available, show key + with args as-is
		b.WriteString(tc.Translate)
		for i := range tc.With {
			write(&tc.With[i], b)
		}
		return
	}

	b.WriteString(tc.Text)
	b.WriteString(tc.Keybind)
	if tc.Score != nil {
		b.WriteString(tc.Score.Name)
	}
	b.WriteString(tc.Selector)
}

// writeFormatted handles MC's Java-style format strings (%s, %1$s, %%, %d).
func writeFormatted(b *strings.Builder, pattern string, args []TextComponent, write componentWriter) {
	seqIdx := 0
	i := 0
	for i < len(pattern) {
		if pattern[i] != '%' || i+1 >= len(pattern) {
			b.WriteByte(pattern[i])
			i++
			continue
		}

		j := i + 1

		// %%
		if pattern[j] == '%' {
			b.WriteByte('%')
			i = j + 1
			continue
		}

		// %N$s (positional)
		if j+2 < len(pattern) && pattern[j] >= '1' && pattern[j] <= '9' && pattern[j+1] == '$' && pattern[j+2] == 's' {
			argIdx := int(pattern[j]-'0') - 1
			if argIdx >= 0 && argIdx < len(args) {
				write(&args[argIdx], b)
			}
			i = j + 3
			continue
		}

		// %s
		if pattern[j] == 's' {
			if seqIdx < len(args) {
				write(&args[seqIdx], b)
				seqIdx++
			}
			i = j + 1
			continue
		}

		// %d
		if pattern[j] == 'd' {
			if seqIdx < len(args) {
				write(&args[seqIdx], b)
				seqIdx++
			}
			i = j + 1
			continue
		}

		// unknown format specifier, output literally
		fmt.Fprintf(b, "%%%c", pattern[j])
		i = j + 1
	}
}

func (tc *TextComponent) writePlain(b *strings.Builder, translate func(string) string) {
	tc.writeContent(b, func(child *TextComponent, b *strings.Builder) {
		child.writePlain(b, translate)
	}, translate)
	for i := range tc.Extra {
		tc.Extra[i].writePlain(b, translate)
	}
}

func (tc *TextComponent) writeMiniMessage(b *strings.Builder) {
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
		for _, arg := range tc.With {
			b.WriteByte(':')
			arg.writeMiniMessage(b)
		}
		b.WriteByte('>')
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

	for _, child := range tc.Extra {
		child.writeMiniMessage(b)
	}

	// close tags in reverse
	for i := len(tags) - 1; i >= 0; i-- {
		b.WriteString("</")
		b.WriteString(tags[i])
		b.WriteByte('>')
	}
}

func (tc *TextComponent) plainTextInto(sb *strings.Builder) {
	sb.WriteString(tc.Text)
	for i := range tc.Extra {
		tc.Extra[i].plainTextInto(sb)
	}
}
