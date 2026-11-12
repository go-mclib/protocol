package wire

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
	r := plainRenderer{translate: translate}
	r.write(&tc)
	return r.b.String()
}

// Translation arguments can be referenced many times (%1$s%1$s), so nested
// components expand exponentially; these limits bound the rendering work.
const (
	maxRenderNodes  = 1 << 16
	maxRenderLength = 1 << 20
)

// plainRenderer renders unformatted text within the render limits.
type plainRenderer struct {
	b         strings.Builder
	translate func(string) string
	nodes     int
}

// write renders tc's content followed by its extras.
func (r *plainRenderer) write(tc *TextComponent) {
	if r.nodes >= maxRenderNodes || r.b.Len() >= maxRenderLength {
		return
	}
	r.nodes++
	r.writeContent(tc)
	for i := range tc.Extra {
		r.write(&tc.Extra[i])
	}
}

// writeContent writes the component's content, without its sibling extras.
func (r *plainRenderer) writeContent(tc *TextComponent) {
	if tc.Translate != "" {
		pattern := tc.Translate
		if tc.Fallback != nil {
			pattern = *tc.Fallback
		}
		if r.translate != nil {
			if resolved := r.translate(tc.Translate); resolved != "" && resolved != tc.Translate {
				pattern = resolved
			}
		}
		r.writeFormatted(pattern, tc.With)
		return
	}
	if tc.Object != nil {
		if tc.Object.Fallback != nil {
			r.write(tc.Object.Fallback)
		} else {
			r.b.WriteString(tc.Object.defaultFallback())
		}
		return
	}

	r.b.WriteString(tc.Text)
	r.b.WriteString(tc.Keybind)
	if tc.Score != nil {
		r.b.WriteString(tc.Score.Name)
	}
	r.b.WriteString(tc.Selector)
}

// writeFormatted follows TranslatableContents: only %s, %N$s and %% are valid.
// Any invalid specifier or missing argument displays the entire original format.
func (r *plainRenderer) writeFormatted(pattern string, args []TextComponent) {
	seq := 0
	for i := 0; i < len(pattern); {
		if pattern[i] != '%' {
			i++
			continue
		}
		end, arg, ok := formatArgument(pattern, i, &seq)
		if !ok || arg >= len(args) {
			r.b.WriteString(pattern)
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
		r.b.WriteString(pattern[start:i])
		end, arg, _ := formatArgument(pattern, i, &seq)
		if arg < 0 {
			r.b.WriteByte('%')
		} else {
			r.write(&args[arg])
		}
		i, start = end, end
	}
	r.b.WriteString(pattern[start:])
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
