package net_structures

import (
	"strconv"
	"strings"
)

// String returns plain text, using translation fallbacks when supplied.
// Without a fallback or resolver, the translation key itself is the format.
func (tc TextComponent) String() string {
	return tc.Render(nil)
}

// Render resolves translation keys with translate. Empty or unchanged keys mean
// missing translations and use Fallback when present, otherwise the key itself.
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

// PlainText returns the same unformatted, fallback-aware content as String.
func (tc *TextComponent) PlainText() string {
	return tc.Render(nil)
}

// componentWriter writes a single component's full content (content + extras) to b.
type componentWriter func(tc *TextComponent, b *strings.Builder)

// writeContent writes the component's content, without its sibling extras.
func (tc *TextComponent) writeContent(b *strings.Builder, write componentWriter, translate func(string) string) {
	if tc.Translate != "" {
		pattern := tc.Translate
		if tc.Fallback != nil {
			pattern = *tc.Fallback
		}
		if translate != nil {
			if resolved := translate(tc.Translate); resolved != "" && resolved != tc.Translate {
				pattern = resolved
			}
		}
		writeFormatted(b, pattern, tc.With, write)
		return
	}
	if tc.Object != nil {
		if tc.Object.Fallback != nil {
			write(tc.Object.Fallback, b)
		} else {
			b.WriteString(tc.Object.defaultFallback())
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

// writeFormatted follows TranslatableContents: only %s, %N$s and %% are valid.
// Any invalid specifier or missing argument displays the entire original format.
func writeFormatted(b *strings.Builder, pattern string, args []TextComponent, write componentWriter) {
	seq := 0
	for i := 0; i < len(pattern); {
		if pattern[i] != '%' {
			i++
			continue
		}
		end, arg, ok := formatArgument(pattern, i, &seq)
		if !ok || arg >= len(args) {
			b.WriteString(pattern)
			return
		}
		i = end
	}

	seq = 0
	start := 0
	for i := 0; i < len(pattern); {
		if pattern[i] != '%' {
			i++
			continue
		}
		b.WriteString(pattern[start:i])
		end, arg, _ := formatArgument(pattern, i, &seq)
		if arg < 0 {
			b.WriteByte('%')
		} else {
			write(&args[arg], b)
		}
		i, start = end, end
	}
	b.WriteString(pattern[start:])
}

// formatArgument returns -1 for literal percent and advances only implicit args.
func formatArgument(pattern string, start int, seq *int) (end, arg int, ok bool) {
	i := start + 1
	if i >= len(pattern) {
		return 0, 0, false
	}
	if pattern[i] == '%' {
		return i + 1, -1, true
	}
	arg = *seq
	digits := i
	for i < len(pattern) && pattern[i] >= '0' && pattern[i] <= '9' {
		i++
	}
	if i > digits {
		if i >= len(pattern) || pattern[i] != '$' {
			return 0, 0, false
		}
		position, err := strconv.ParseInt(pattern[digits:i], 10, 32)
		if err != nil || position <= 0 {
			return 0, 0, false
		}
		arg = int(position) - 1
		i++
	} else {
		*seq++
	}
	if i >= len(pattern) || pattern[i] != 's' {
		return 0, 0, false
	}
	return i + 1, arg, true
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
	} else if tc.Object != nil {
		if tc.Object.Fallback != nil {
			tc.Object.Fallback.writeMiniMessage(b)
		} else {
			b.WriteString(tc.Object.defaultFallback())
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
