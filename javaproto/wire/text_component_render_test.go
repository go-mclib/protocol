package wire_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/go-mclib/protocol/javaproto/textformat"
	ns "github.com/go-mclib/protocol/javaproto/wire"
)

func boolPtr(v bool) *bool { return &v }

func TestTextComponent_String(t *testing.T) {
	cases := []struct {
		name string
		tc   ns.TextComponent
		want string
	}{
		{"plain", ns.TextComponent{Text: "Hello"}, "Hello"},
		{"with extra", ns.TextComponent{Text: "Hello, ", Extra: []ns.TextComponent{{Text: "World"}}}, "Hello, World"},
		{"translate", ns.TextComponent{Translate: "chat.type.text"}, "chat.type.text"},
		{"unresolved translation excludes unused args", ns.TextComponent{Translate: "chat.type.text", With: []ns.TextComponent{{Text: "Player"}, {Text: "Hello"}}}, "chat.type.text"},
		{"nested", ns.TextComponent{Text: "a", Extra: []ns.TextComponent{{Text: "b", Extra: []ns.TextComponent{{Text: "c"}}}}}, "abc"},
		{"empty", ns.TextComponent{}, ""},
		{"keybind", ns.TextComponent{Keybind: "key.jump"}, "key.jump"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := c.tc.String()
			if got != c.want {
				t.Errorf("String() = %q, want %q", got, c.want)
			}
		})
	}
}

func TestTextComponent_UnmarshalJSON(t *testing.T) {
	cases := []struct {
		name string
		json string
		want string
	}{
		{"plain string", `"Hello"`, "Hello"},
		{"object with text", `{"text":"Hello"}`, "Hello"},
		{"with color", `{"text":"Hello","color":"red"}`, "Hello"},
		{"with extra", `{"text":"Hello ","extra":[{"text":"World"}]}`, "Hello World"},
		{"translate", `{"translate":"chat.type.text","with":[{"text":"Player"},{"text":"msg"}]}`, "chat.type.text"},
		{"nested extra", `{"text":"a","extra":[{"text":"b","extra":[{"text":"c"}]}]}`, "abc"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var tc ns.TextComponent
			if err := json.Unmarshal([]byte(c.json), &tc); err != nil {
				t.Fatalf("Unmarshal(%s) error: %v", c.json, err)
			}
			got := tc.String()
			if got != c.want {
				t.Errorf("String() = %q, want %q", got, c.want)
			}
		})
	}
}

func TestTextComponent_TranslationFallback(t *testing.T) {
	const raw = `{"translate":"plugin.discord","fallback":"[Discord] %2$s » %1$s","with":[["A",{"text":"aaaaa"}],{"text":"Era","extra":["gon"]}],"extra":["!"]}`
	var tc ns.TextComponent
	if err := json.Unmarshal([]byte(raw), &tc); err != nil {
		t.Fatal(err)
	}
	const want = "[Discord] Eragon » Aaaaaa!"
	for name, got := range map[string]string{
		"String":            tc.String(),
		"missing resolver":  tc.Render(func(string) string { return "" }),
		"identity resolver": tc.Render(func(key string) string { return key }),
	} {
		if got != want {
			t.Errorf("%s = %q, want %q", name, got, want)
		}
	}
	if got := tc.Render(func(string) string { return "%1$s from %2$s" }); got != "Aaaaaa from Eragon!" {
		t.Errorf("resolved translation = %q", got)
	}
	if err := json.Unmarshal([]byte(`{"translate":"plugin.empty","fallback":"","with":["hidden"]}`), &tc); err != nil {
		t.Fatal(err)
	}
	if got := tc.String(); got != "" {
		t.Errorf("explicit empty fallback = %q", got)
	}
}

func TestTextComponent_TranslationFormat(t *testing.T) {
	args := []ns.TextComponent{
		{Text: "one"}, {Text: "two"}, {Text: "three"}, {Text: "four"}, {Text: "five"},
		{Text: "six"}, {Text: "seven"}, {Text: "eight"}, {Text: "nine"}, {Text: "ten"},
	}
	for _, test := range []struct{ pattern, want string }{
		{"%10$s/%s/%2$s/%s/%%", "ten/one/two/two/%"},
		{"%01$s", "one"},
		{"%s then %11$s", "%s then %11$s"},
		{"%s then %d", "%s then %d"},
		{"%0$s", "%0$s"},
		{"%2147483648$s", "%2147483648$s"},
		{"%s trailing %", "%s trailing %"},
		{"%s then %1$%", "%s then %1$%"},
		{"%s then %?", "%s then %?"},
	} {
		t.Run(test.pattern, func(t *testing.T) {
			tc := ns.TextComponent{Translate: "plugin.format", With: args, Extra: []ns.TextComponent{{Text: "!"}}}
			if got := tc.Render(func(string) string { return test.pattern }); got != test.want+"!" {
				t.Errorf("Render = %q, want %q", got, test.want+"!")
			}
		})
	}
}

func TestTextComponent_JSONLists(t *testing.T) {
	const raw = `[{"text":"[","color":"blue","bold":true,"extra":["Dis"]},["cord",{"text":"] ","color":"red"}],"Eragon",{"text":" » Aaaaaa"}]`
	var tc ns.TextComponent
	if err := json.Unmarshal([]byte(raw), &tc); err != nil {
		t.Fatal(err)
	}
	if got := tc.String(); got != "[Discord] Eragon » Aaaaaa" {
		t.Errorf("String = %q", got)
	}
	if got := textformat.MiniMessage(tc); got != "<blue><bold>[Discord<red>] </red>Eragon » Aaaaaa</bold></blue>" {
		t.Errorf("list style inheritance = %q", got)
	}
	if err := json.Unmarshal([]byte(`{"translate":"plugin.count","fallback":"%s %s","with":[42,true]}`), &tc); err != nil {
		t.Fatal(err)
	}
	if got := tc.String(); got != "42 true" {
		t.Errorf("primitive arguments = %q", got)
	}
}

func TestTextComponent_RejectMalformedJSON(t *testing.T) {
	for _, raw := range []string{
		`null`, `true`, `42`, `[]`, `{}`, `{"text":null}`, `{"text":42}`,
		`["[",null]`, `["[",[]]`, `{"text":"","extra":[]}`,
		`{"text":"","extra":[{}]}`, `{"text":"","extra":null}`,
		`{"translate":"x","with":[null]}`, `{"translate":"x","with":null}`,
		`{"type":"unknown","text":"hidden"}`, `{"type":"translatable","text":"hidden"}`,
	} {
		t.Run(raw, func(t *testing.T) {
			tc := ns.TextComponent{Text: "unchanged"}
			if err := json.Unmarshal([]byte(raw), &tc); err == nil {
				t.Fatalf("accepted malformed component %s", raw)
			}
			if tc.String() != "unchanged" {
				t.Errorf("failed decode replaced previous component: %q", tc.String())
			}
		})
	}
}

func TestRenderBoundsExponentialTranslation(t *testing.T) {
	// each level references its argument twice: 2^60 expansions if unbounded
	tc := ns.NewTextComponent("x")
	for range 60 {
		tc = ns.NewTranslateComponent("%1$s%1$s", tc)
	}
	done := make(chan string, 1)
	go func() { done <- tc.String() }()
	select {
	case out := <-done:
		if len(out) == 0 || len(out) > 2<<20 {
			t.Fatalf("rendered %d bytes", len(out))
		}
	case <-time.After(5 * time.Second):
		t.Fatal("rendering did not terminate")
	}
}
