package auth

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	msAuthorizeURL = "https://login.live.com/oauth20_authorize.srf"
	msTokenURL     = "https://login.live.com/oauth20_token.srf"
)

func buildAuthorizeURL(clientID, redirectURL, state string, scopes []string) string {
	q := url.Values{}
	q.Set("client_id", clientID)
	q.Set("response_type", "code")
	q.Set("redirect_uri", redirectURL)
	q.Set("scope", strings.Join(scopes, " "))
	q.Set("prompt", "select_account")
	q.Set("state", state)
	return msAuthorizeURL + "?" + q.Encode()
}

// newOAuthState returns a random value binding the callback to this login attempt.
func newOAuthState() string {
	return rand.Text()
}

// startLocalServer serves the OAuth callback on 127.0.0.1:port (0 picks a free
// port) and returns the bound port. Callbacks without the expected state are
// ignored, so stray local requests can neither inject a code nor abort the flow.
func startLocalServer(port int, state string, codeCh chan<- string) (*http.Server, int, error) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		values := r.URL.Query()
		if values.Get("state") != state {
			http.NotFound(w, r)
			return
		}
		code := values.Get("code")
		if code == "" {
			_, _ = io.WriteString(w, "Cannot authenticate.")
		} else {
			_, _ = io.WriteString(w, "You may now close this page.")
		}
		select {
		case codeCh <- code: // empty code reports a failed authorization
		default: // prevent deadlock when channel is full
		}
	})

	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return nil, 0, err
	}
	srv := &http.Server{Handler: mux}
	go func() {
		_ = srv.Serve(ln)
	}()
	return srv, ln.Addr().(*net.TCPAddr).Port, nil
}

func stopLocalServer(srv *http.Server) error {
	if srv == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return srv.Shutdown(ctx)
}

type msTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// requestMSToken calls the Microsoft token endpoint; form carries the grant
// (authorization_code + code, or refresh_token + refresh_token).
func requestMSToken(ctx context.Context, httpClient *http.Client, clientID, redirectURL string, form url.Values, op string) (*msTokenResponse, error) {
	form.Set("client_id", clientID)
	form.Set("redirect_uri", redirectURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, msTokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	var tr msTokenResponse
	if err := doJSON(httpClient, req, op, &tr); err != nil {
		return nil, err
	}
	return &tr, nil
}
