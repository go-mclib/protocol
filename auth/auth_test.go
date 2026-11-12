package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	ns "github.com/go-mclib/protocol/javaproto/wire"
)

// routeTransport serves canned JSON bodies keyed by request URL (without query).
type routeTransport map[string]string

func (rt routeTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	u := *req.URL
	u.RawQuery = ""
	body, ok := rt[u.String()]
	status := http.StatusOK
	if !ok {
		status, body = http.StatusNotFound, "{}"
	}
	return &http.Response{
		StatusCode: status,
		Status:     http.StatusText(status),
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     http.Header{"Content-Type": {"application/json"}},
		Request:    req,
	}, nil
}

func loginRoutes(xbl, msToken string) routeTransport {
	return routeTransport{
		msTokenURL:             msToken,
		xblUserAuthenticateURL: xbl,
		xstsAuthorizeURL:       `{"Token":"xsts"}`,
		mcAuthLoginWithXboxURL: `{"access_token":"mc","expires_in":86400}`,
		mcEntitlementsURL:      `{"items":[{"name":"product_minecraft"},{"name":"game_minecraft"}]}`,
		mcProfileURL:           `{"id":"069a79f444e94726a5befca90e38aaf5","name":"Notch"}`,
	}
}

func newTestClient(rt routeTransport) *Client {
	return NewClient(Config{
		ClientID:   "client",
		HTTPClient: &http.Client{Transport: rt},
		TokenStore: NewMemoryTokenStore(),
	})
}

func TestLoginWithRefreshTokenRejectsEmptyUserHash(t *testing.T) {
	c := newTestClient(loginRoutes(`{"Token":"xbl","DisplayClaims":{"xui":[]}}`, `{"access_token":"ms","refresh_token":"new"}`))
	if _, err := c.LoginWithRefreshToken(context.Background(), "old"); err == nil {
		t.Fatal("expected error for XBL response without user hash")
	}
}

func TestLoginWithRefreshTokenKeepsTokenWhenNotRotated(t *testing.T) {
	c := newTestClient(loginRoutes(`{"Token":"xbl","DisplayClaims":{"xui":[{"uhs":"hash"}]}}`, `{"access_token":"ms"}`))
	data, err := c.LoginWithRefreshToken(context.Background(), "old")
	if err != nil {
		t.Fatal(err)
	}
	notch, _ := ns.UUIDFromString("069a79f4-44e9-4726-a5be-fca90e38aaf5")
	if data.RefreshToken != "old" || data.Username != "Notch" || data.AccessToken != "mc" || data.UUID != notch {
		t.Fatalf("unexpected login data %+v", data)
	}
}

func TestLocalServerIgnoresForeignState(t *testing.T) {
	codeCh := make(chan string, 1)
	srv, port, err := startLocalServer(0, "expected", codeCh)
	if err != nil {
		t.Fatal(err)
	}
	defer stopLocalServer(srv)

	get := func(query string) int {
		res, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/?%s", port, query))
		if err != nil {
			t.Fatal(err)
		}
		res.Body.Close()
		return res.StatusCode
	}
	if status := get("code=evil&state=wrong"); status != http.StatusNotFound {
		t.Fatalf("foreign state answered %d", status)
	}
	if status := get("code=good&state=expected"); status != http.StatusOK {
		t.Fatalf("valid callback answered %d", status)
	}
	select {
	case code := <-codeCh:
		if code != "good" {
			t.Fatalf("got code %q", code)
		}
	case <-time.After(time.Second):
		t.Fatal("no code delivered")
	}
}

func TestFileTokenStoreRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "creds.json")
	store, err := NewFileTokenStore(path)
	if err != nil {
		t.Fatal(err)
	}
	want := &LoginData{AccessToken: "a", RefreshToken: "r", UUID: ns.UUID{1, 2, 3}, Username: "Steve", ExpiresAt: time.Now().Add(time.Hour)}
	if err := store.SaveSession(want); err != nil {
		t.Fatal(err)
	}
	got, err := store.LoadSession("Steve")
	if err != nil || got == nil || !got.ExpiresAt.Equal(want.ExpiresAt) ||
		got.AccessToken != want.AccessToken || got.RefreshToken != want.RefreshToken || got.UUID != want.UUID {
		t.Fatalf("got %+v, %v; want %+v", got, err, want)
	}
	if !got.IsValid() {
		t.Fatal("fresh session should be valid")
	}
	if matches, _ := filepath.Glob(filepath.Join(filepath.Dir(path), ".credentials-*.tmp")); len(matches) != 0 {
		t.Fatalf("temp files left behind: %v", matches)
	}
}

func TestFileTokenStoreLoadsOlderCaches(t *testing.T) {
	notch, _ := ns.UUIDFromString("069a79f4-44e9-4726-a5be-fca90e38aaf5")
	for name, test := range map[string]struct {
		file string
		want LoginData
	}{
		// sessions cached before LoginData.UUID was typed hold Mojang's dashless id
		"dashless uuid": {
			`{"Notch":{"accessToken":"a","refreshToken":"r","uuid":"069a79f444e94726a5befca90e38aaf5","username":"Notch","expiresAt":"2030-01-02T03:04:05Z"}}`,
			LoginData{AccessToken: "a", RefreshToken: "r", UUID: notch, Username: "Notch", ExpiresAt: time.Date(2030, 1, 2, 3, 4, 5, 0, time.UTC)},
		},
		"dashed uuid": {
			`{"Notch":{"uuid":"069a79f4-44e9-4726-a5be-fca90e38aaf5","username":"Notch"}}`,
			LoginData{UUID: notch, Username: "Notch"},
		},
		// entries migrated from the refresh-token-only format were saved without a uuid
		"empty uuid": {
			`{"Notch":{"accessToken":"","refreshToken":"r","uuid":"","username":"Notch","expiresAt":"0001-01-01T00:00:00Z"}}`,
			LoginData{RefreshToken: "r", Username: "Notch"},
		},
		"refresh-token-only format": {
			`{"Notch":"r"}`,
			LoginData{RefreshToken: "r", Username: "Notch"},
		},
	} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "creds.json")
			if err := os.WriteFile(path, []byte(test.file), 0o600); err != nil {
				t.Fatal(err)
			}
			store, err := NewFileTokenStore(path)
			if err != nil {
				t.Fatal(err)
			}
			got, err := store.LoadSession("Notch")
			if err != nil || got == nil || got.UUID != test.want.UUID || got.RefreshToken != test.want.RefreshToken ||
				got.AccessToken != test.want.AccessToken || got.Username != test.want.Username || !got.ExpiresAt.Equal(test.want.ExpiresAt) {
				t.Fatalf("got %+v, %v; want %+v", got, err, test.want)
			}
		})
	}
}

func TestLoginDataRejectsMalformedUUID(t *testing.T) {
	var d LoginData
	if err := json.Unmarshal([]byte(`{"uuid":"not-a-uuid"}`), &d); err == nil {
		t.Fatal("accepted malformed uuid")
	}
}

func TestLoginDataJSONUsesDashedUUID(t *testing.T) {
	data, err := json.Marshal(LoginData{UUID: ns.UUID{15: 1}})
	if err != nil || !strings.Contains(string(data), `"uuid":"00000000-0000-0000-0000-000000000001"`) {
		t.Fatalf("marshal = %s, %v", data, err)
	}
}

func TestFileTokenStoreRejectsReadOnlyFile(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root bypasses file permissions")
	}
	path := filepath.Join(t.TempDir(), "creds.json")
	if err := os.WriteFile(path, []byte("{}"), 0o400); err != nil {
		t.Fatal(err)
	}
	if _, err := NewFileTokenStore(path); err == nil {
		t.Fatal("accepted a read-only credentials file")
	}
}
