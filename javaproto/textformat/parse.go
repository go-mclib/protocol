// Package textformat converts between text components and human-authored
// formatting: legacy color codes (§/&) and Adventure MiniMessage tags.
package textformat

import (
	"strings"

	ns "github.com/go-mclib/protocol/javaproto/wire"
)

// color code mappings (§ and & codes)
var colorCodes = map[byte]string{
	'0': "black",
	'1': "dark_blue",
	'2': "dark_green",
	'3': "dark_aqua",
	'4': "dark_red",
	'5': "dark_purple",
	'6': "gold",
	'7': "gray",
	'8': "dark_gray",
	'9': "blue",
	'a': "green",
	'b': "aqua",
	'c': "red",
	'd': "light_purple",
	'e': "yellow",
	'f': "white",
}

func boolPtr(v bool) *bool { return &v }

// ParseFormatted parses a string containing legacy color codes (§ or &)
// and MiniMessage-style tags (<red>, <bold>, <gradient:red:blue>, etc.)
// into a text component tree.
func ParseFormatted(s string) ns.TextComponent {
	// first pass: normalize § to & for uniform processing
	s = strings.ReplaceAll(s, "§", "&")

	// check if string has any formatting
	if !strings.ContainsAny(s, "&<") {
		return ns.TextComponent{Text: s}
	}

	// parse MiniMessage tags first, then legacy codes within each segment
	root := ns.TextComponent{Text: ""}
	parseMixed(s, &root)

	if len(root.Extra) == 1 && root.Text == "" {
		return root.Extra[0]
	}
	return root
}

// parseMixed handles a string that may contain both <tags> and &codes.
func parseMixed(s string, parent *ns.TextComponent) {
	for len(s) > 0 {
		// find next tag
		tagStart := strings.Index(s, "<")
		if tagStart == -1 {
			// no more tags, parse remaining as legacy
			parseLegacyCodes(s, parent)
			return
		}

		// parse legacy codes before the tag
		if tagStart > 0 {
			parseLegacyCodes(s[:tagStart], parent)
		}

		// find closing >
		tagEnd := strings.Index(s[tagStart:], ">")
		if tagEnd == -1 {
			// malformed tag, treat rest as literal
			parseLegacyCodes(s[tagStart:], parent)
			return
		}
		tagEnd += tagStart

		tagContent := s[tagStart+1 : tagEnd]
		s = s[tagEnd+1:]

		// check if it's a closing tag
		if strings.HasPrefix(tagContent, "/") {
			// closing tag — return to parent (handled by recursion)
			continue
		}

		// find the matching closing tag; parameterized tags like <gradient:red:blue> close by name
		baseTag, _, _ := strings.Cut(tagContent, ":")
		inner, rest, _ := strings.Cut(s, "</"+baseTag+">") // no closing tag: rest of string is content
		s = rest

		child := applyTag(tagContent)
		parseMixed(inner, &child)
		parent.Extra = append(parent.Extra, child)
	}
}

// applyTag creates a component with the style for a tag name.
func applyTag(tag string) ns.TextComponent {
	tc := ns.TextComponent{Text: ""}

	// handle parameterized tags
	name, arg, hasArg := strings.Cut(tag, ":")
	name = strings.ToLower(name)

	switch name {
	// formatting
	case "bold", "b":
		tc.Bold = boolPtr(true)
	case "italic", "i", "em":
		tc.Italic = boolPtr(true)
	case "underlined", "underline", "u":
		tc.Underlined = boolPtr(true)
	case "strikethrough", "st":
		tc.Strikethrough = boolPtr(true)
	case "obfuscated", "obf", "magic":
		tc.Obfuscated = boolPtr(true)

	// hex color: <#ff5555> or <color:#ff5555>
	case "color":
		if hasArg {
			tc.Color = arg
		}

	// gradient (simplified: use first color)
	case "gradient":
		if hasArg {
			tc.Color, _, _ = strings.Cut(arg, ":")
		}

	// reset
	case "reset":
		tc.Color = "white"
		tc.Bold = boolPtr(false)
		tc.Italic = boolPtr(false)
		tc.Underlined = boolPtr(false)
		tc.Strikethrough = boolPtr(false)
		tc.Obfuscated = boolPtr(false)

	default:
		// named color (<red>) or hex (<#ff5555>)
		if isNamedColor(name) || strings.HasPrefix(name, "#") {
			tc.Color = name
		}
	}

	return tc
}

func isNamedColor(name string) bool {
	for _, color := range colorCodes {
		if color == name {
			return true
		}
	}
	return false
}

// legacyStyle returns the style field set by a legacy formatting code, or nil.
func legacyStyle(tc *ns.TextComponent, code byte) **bool {
	switch code {
	case 'l':
		return &tc.Bold
	case 'o':
		return &tc.Italic
	case 'n':
		return &tc.Underlined
	case 'm':
		return &tc.Strikethrough
	case 'k':
		return &tc.Obfuscated
	}
	return nil
}

// parseLegacyCodes parses &-codes in a string and appends segments to parent.
func parseLegacyCodes(s string, parent *ns.TextComponent) {
	if !strings.Contains(s, "&") {
		if s != "" {
			parent.Extra = append(parent.Extra, ns.TextComponent{Text: s})
		}
		return
	}

	var current ns.TextComponent
	flush := func() {
		if current.Text != "" {
			parent.Extra = append(parent.Extra, current)
		}
	}
	for i := 0; i < len(s); {
		if s[i] == '&' && i+1 < len(s) {
			code := toLower(s[i+1])
			if color, ok := colorCodes[code]; ok {
				flush()
				current = ns.TextComponent{Color: color}
				i += 2
				continue
			}
			if code == 'r' {
				flush()
				current = ns.TextComponent{}
				i += 2
				continue
			}
			if legacyStyle(&current, code) != nil {
				// styling after text starts a new segment that keeps only the color
				if current.Text != "" {
					flush()
					current = ns.TextComponent{Color: current.Color}
				}
				*legacyStyle(&current, code) = boolPtr(true)
				i += 2
				continue
			}
		}

		// copy the literal run up to the next '&' (byte-wise, so UTF-8 stays intact)
		end := len(s)
		if j := strings.IndexByte(s[i+1:], '&'); j >= 0 {
			end = i + 1 + j
		}
		current.Text += s[i:end]
		i = end
	}
	flush()
}

func toLower(b byte) byte {
	if b >= 'A' && b <= 'Z' {
		return b + 32
	}
	return b
}
