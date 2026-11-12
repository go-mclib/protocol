package textformat_test

import (
	"testing"

	"github.com/go-mclib/protocol/javaproto/textformat"
	ns "github.com/go-mclib/protocol/javaproto/wire"
)

func boolPtr(v bool) *bool { return &v }

func TestMiniMessage(t *testing.T) {
	tc := ns.TextComponent{Text: "Hello", Color: "red"}
	got := textformat.MiniMessage(tc)
	if got != "<red>Hello</red>" {
		t.Errorf("MiniMessage() = %q, want %q", got, "<red>Hello</red>")
	}

	// translate
	tc = ns.TextComponent{
		Translate: "chat.type.text",
		With:      []ns.TextComponent{{Text: "Player"}, {Text: "Hello"}},
	}
	got = textformat.MiniMessage(tc)
	if got != "<lang:chat.type.text:Player:Hello>" {
		t.Errorf("MiniMessage() = %q, want %q", got, "<lang:chat.type.text:Player:Hello>")
	}

	// bold + color
	tc = ns.TextComponent{Text: "wow", Color: "gold", Bold: boolPtr(true)}
	got = textformat.MiniMessage(tc)
	if got != "<gold><bold>wow</bold></gold>" {
		t.Errorf("MiniMessage() = %q, want %q", got, "<gold><bold>wow</bold></gold>")
	}
}
