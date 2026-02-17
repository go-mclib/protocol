package net_structures

import (
	"fmt"
	"strings"
)

// TODO: support hex color coded, add "From" methods, e.g. FromMiniMessage

// MC color name -> ANSI escape code
var mcColorToANSI = map[string]string{
	"black":        "\033[30m",
	"dark_blue":    "\033[34m",
	"dark_green":   "\033[32m",
	"dark_aqua":    "\033[36m",
	"dark_red":     "\033[31m",
	"dark_purple":  "\033[35m",
	"gold":         "\033[33m",
	"gray":         "\033[37m",
	"dark_gray":    "\033[90m",
	"blue":         "\033[94m",
	"green":        "\033[92m",
	"aqua":         "\033[96m",
	"red":          "\033[91m",
	"light_purple": "\033[95m",
	"yellow":       "\033[93m",
	"white":        "\033[97m",
}

// MC color name - Bukkit section code
var mcColorToCode = map[string]string{
	"black":        "§0",
	"dark_blue":    "§1",
	"dark_green":   "§2",
	"dark_aqua":    "§3",
	"dark_red":     "§4",
	"dark_purple":  "§5",
	"gold":         "§6",
	"gray":         "§7",
	"dark_gray":    "§8",
	"blue":         "§9",
	"green":        "§a",
	"aqua":         "§b",
	"red":          "§c",
	"light_purple": "§d",
	"yellow":       "§e",
	"white":        "§f",
}

// String returns the plain text content of the component and all children,
// with no formatting. Translate keys are included as-is.
func (tc TextComponent) String() string {
	var b strings.Builder
	tc.writePlain(&b)
	return b.String()
}

func (tc *TextComponent) writePlain(b *strings.Builder) {
	b.WriteString(tc.Text)
	b.WriteString(tc.Translate)
	b.WriteString(tc.Keybind)
	if tc.Score != nil {
		b.WriteString(tc.Score.Name)
	}
	b.WriteString(tc.Selector)

	for _, child := range tc.With {
		child.writePlain(b)
	}
	for _, child := range tc.Extra {
		child.writePlain(b)
	}
}

// ANSI returns the text with ANSI terminal escape codes for colors and formatting.
func (tc TextComponent) ANSI() string {
	var b strings.Builder
	if tc.writeANSI(&b) {
		b.WriteString("\033[0m")
	}
	return b.String()
}

func (tc *TextComponent) writeANSI(b *strings.Builder) bool {
	prefix := tc.ansiPrefix()
	styled := prefix != ""
	if styled {
		b.WriteString(prefix)
	}

	b.WriteString(tc.Text)
	b.WriteString(tc.Translate)
	b.WriteString(tc.Keybind)
	if tc.Score != nil {
		b.WriteString(tc.Score.Name)
	}
	b.WriteString(tc.Selector)

	for _, child := range tc.With {
		if child.writeANSI(b) {
			styled = true
		}
	}
	for _, child := range tc.Extra {
		// reset before each styled child so parent style doesn't bleed
		if styled {
			b.WriteString("\033[0m")
		}
		if child.writeANSI(b) {
			styled = true
		}
	}
	return styled
}

func (tc *TextComponent) ansiPrefix() string {
	var codes []string

	if tc.Color != "" {
		if ansi, ok := mcColorToANSI[tc.Color]; ok {
			codes = append(codes, ansi)
		} else if strings.HasPrefix(tc.Color, "#") && len(tc.Color) == 7 {
			// hex color → 24-bit ANSI
			var r, g, b int
			fmt.Sscanf(tc.Color[1:], "%02x%02x%02x", &r, &g, &b)
			codes = append(codes, fmt.Sprintf("\033[38;2;%d;%d;%dm", r, g, b))
		}
	}
	if tc.Bold != nil && *tc.Bold {
		codes = append(codes, "\033[1m")
	}
	if tc.Italic != nil && *tc.Italic {
		codes = append(codes, "\033[3m")
	}
	if tc.Underlined != nil && *tc.Underlined {
		codes = append(codes, "\033[4m")
	}
	if tc.Strikethrough != nil && *tc.Strikethrough {
		codes = append(codes, "\033[9m")
	}
	if tc.Obfuscated != nil && *tc.Obfuscated {
		codes = append(codes, "\033[8m")
	}

	return strings.Join(codes, "")
}

// ColorCodes returns the text with Bukkit-style section sign (§) color codes.
func (tc TextComponent) ColorCodes() string {
	var b strings.Builder
	tc.writeColorCodes(&b)
	return b.String()
}

func (tc *TextComponent) writeColorCodes(b *strings.Builder) {
	if tc.Color != "" {
		if code, ok := mcColorToCode[tc.Color]; ok {
			b.WriteString(code)
		}
	}
	if tc.Bold != nil && *tc.Bold {
		b.WriteString("§l")
	}
	if tc.Italic != nil && *tc.Italic {
		b.WriteString("§o")
	}
	if tc.Underlined != nil && *tc.Underlined {
		b.WriteString("§n")
	}
	if tc.Strikethrough != nil && *tc.Strikethrough {
		b.WriteString("§m")
	}
	if tc.Obfuscated != nil && *tc.Obfuscated {
		b.WriteString("§k")
	}

	b.WriteString(tc.Text)
	b.WriteString(tc.Translate)
	b.WriteString(tc.Keybind)
	if tc.Score != nil {
		b.WriteString(tc.Score.Name)
	}
	b.WriteString(tc.Selector)

	for _, child := range tc.With {
		child.writeColorCodes(b)
	}
	for _, child := range tc.Extra {
		child.writeColorCodes(b)
	}
}

// MiniMessage returns the text in Adventure MiniMessage format.
func (tc TextComponent) MiniMessage() string {
	var b strings.Builder
	tc.writeMiniMessage(&b)
	return b.String()
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
