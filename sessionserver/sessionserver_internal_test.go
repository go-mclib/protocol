package sessionserver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	ns "github.com/go-mclib/protocol/javaproto/wire"
)

func TestServerHash(t *testing.T) {
	// https://github.com/PrismarineJS/node-yggdrasil/blob/c2b1e534dc56d33d8ea0c1ba02ead058b9db07b1/test/index.js#L70
	if got, want := computeServerHash("cat", []byte("cat"), []byte("cat")), "-af59e5b1d5d92e5c2c2776ed0e65e90be181f2a"; got != want {
		t.Errorf("computeServerHash() = %q, want %q", got, want)
	}
}

func TestValidateAccessToken(t *testing.T) {
	tests := []struct {
		token    string
		expected bool
	}{
		{"validtoken123456", true},
		{"a_very_long_access_token_that_is_longer_than_usual_access_token_but_should_still_be_valid", true},
		{"", false},                         // empty
		{"short", false},                    // too short
		{string(make([]byte, 3000)), false}, // too long
	}

	for _, test := range tests {
		if result := validateAccessToken(test.token); result != test.expected {
			t.Errorf("validateAccessToken(%q) = %v, expected %v", test.token, result, test.expected)
		}
	}
}

func TestHasJoinedRequest(t *testing.T) {
	secret, key := []byte("cat"), []byte("cat")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if r.Method != http.MethodGet || r.URL.Path != "/session/minecraft/hasJoined" || r.Header.Get("User-Agent") != userAgent ||
			q.Get("username") != "a&b c" || q.Get("serverId") != computeServerHash("", secret, key) || q.Get("ip") != "::1" || len(q) != 3 {
			http.Error(w, `{"error":"bad request"}`, http.StatusBadRequest)
			return
		}
		_, _ = w.Write([]byte(`{"id":"00000000000000000000000000000001","name":"a&b c","properties":[{"name":"textures","value":"v","signature":"s"}]}`))
	}))
	defer srv.Close()

	c := New(srv.Client())
	c.BaseURL = srv.URL
	resp, err := c.HasJoined(context.Background(), "a&b c", "", secret, key, "::1")
	if err != nil {
		t.Fatal(err)
	}
	if resp == nil || resp.Name != "a&b c" || resp.ID != (ns.UUID{15: 1}) || len(resp.Properties) != 1 || resp.Properties[0].Signature != "s" {
		t.Fatalf("unexpected response %+v", resp)
	}
}

func TestHasJoinedNotJoined(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Has("ip") {
			http.Error(w, `{"error":"unexpected ip"}`, http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c := New(srv.Client())
	c.BaseURL = srv.URL
	resp, err := c.HasJoined(context.Background(), "Steve", "", nil, nil, "")
	if err != nil || resp != nil {
		t.Fatalf("got %+v, %v; want nil, nil", resp, err)
	}
}

func TestHasJoinedRejectsMalformedID(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"id":"not-a-uuid","name":"Steve"}`))
	}))
	defer srv.Close()

	c := New(srv.Client())
	c.BaseURL = srv.URL
	if _, err := c.HasJoined(context.Background(), "Steve", "", nil, nil, ""); err == nil {
		t.Fatal("accepted malformed profile id")
	}
}

func TestJoinRequest(t *testing.T) {
	profile := ns.UUID{0x55, 0x0e, 0x84, 0x00, 0xe2, 0x9b, 0x41, 0xd4, 0xa7, 0x16, 0x44, 0x66, 0x55, 0x44, 0x00, 0x00}
	secret, key := []byte("secret"), []byte("key")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var got joinRequest
		if r.Method != http.MethodPost || r.URL.Path != "/session/minecraft/join" || r.Header.Get("Content-Type") != "application/json" ||
			json.NewDecoder(r.Body).Decode(&got) != nil {
			http.Error(w, `{"error":"bad request"}`, http.StatusBadRequest)
			return
		}
		want := joinRequest{AccessToken: "validtoken123456", SelectedProfile: "550e8400e29b41d4a716446655440000", ServerID: computeServerHash("id", secret, key)}
		if got != want {
			http.Error(w, `{"error":"bad body"}`, http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c := New(srv.Client())
	c.BaseURL = srv.URL
	if err := c.Join(context.Background(), "validtoken123456", profile, "id", secret, key); err != nil {
		t.Fatal(err)
	}
}

func TestJoinReportsAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":"ForbiddenOperationException","errorMessage":"Invalid token","path":"/session/minecraft/join"}`))
	}))
	defer srv.Close()

	c := New(srv.Client())
	c.BaseURL = srv.URL
	if err := c.Join(context.Background(), "validtoken123456", ns.UUID{1}, "", nil, nil); err == nil {
		t.Fatal("Join succeeded on 403")
	}
}

func TestJoinHonorsContext(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	c := New(srv.Client())
	c.BaseURL = srv.URL
	if err := c.Join(ctx, "validtoken123456", ns.UUID{1}, "", nil, nil); err == nil {
		t.Fatal("Join succeeded with a cancelled context")
	}
}

func TestNewDefaults(t *testing.T) {
	c := New(nil)
	if c.BaseURL != defaultBaseURL || c.httpClient == nil || c.httpClient.Timeout == 0 {
		t.Fatalf("unexpected defaults: %+v", c)
	}
}
