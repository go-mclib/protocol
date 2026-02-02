package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	browser "github.com/pkg/browser"
)

const (
	msAuthorizeURL = "https://login.live.com/oauth20_authorize.srf"
	msTokenURL     = "https://login.live.com/oauth20_token.srf"
)

func buildAuthorizeURL(clientID, redirectURL string, scopes []string) string {
	q := url.Values{}
	q.Set("client_id", clientID)
	q.Set("response_type", "code")
	q.Set("redirect_uri", redirectURL)
	q.Set("scope", strings.Join(scopes, " "))
	q.Set("prompt", "select_account")
	return msAuthorizeURL + "?" + q.Encode()
}

func startLocalServer(port int, codeCh chan<- string) (*http.Server, error) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		values := r.URL.Query()
		code := values.Get("code")
		if code == "" {
			_, _ = io.WriteString(w, "Cannot authenticate.")
			select {
			case codeCh <- "": // no code
			default: // prevent deadlock when channel is full
			}
			return
		}
		_, _ = io.WriteString(w, "You may now close this page.")
		select {
		case codeCh <- code: // send code
		default:
		}
	})

	srv := &http.Server{Handler: mux}
	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return nil, err
	}

	go func() {
		_ = srv.Serve(ln)
	}()
	return srv, nil
}

func stopLocalServer(srv *http.Server) error {
	if srv == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return srv.Shutdown(ctx)
}

func openBrowser(url string) error {
	return browser.OpenURL(url)
}

type msTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

func exchangeAuthCodeForTokens(ctx context.Context, httpClient *http.Client, clientID, redirectURL, code string) (*msTokenResponse, error) {
	form := url.Values{}
	form.Set("client_id", clientID)
	form.Set("code", code)
	form.Set("grant_type", "authorization_code")
	form.Set("redirect_uri", redirectURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, msTokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		data, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("token exchange failed: %s: %s", res.Status, string(data))
	}

	var tr msTokenResponse
	if err := json.NewDecoder(res.Body).Decode(&tr); err != nil {
		return nil, err
	}

	return &tr, nil
}

func refreshAccessToken(ctx context.Context, httpClient *http.Client, clientID, redirectURL, refreshToken string) (*msTokenResponse, error) {
	if refreshToken == "" {
		return nil, errors.New("empty refresh token")
	}
	form := url.Values{}
	form.Set("client_id", clientID)
	form.Set("refresh_token", refreshToken)
	form.Set("grant_type", "refresh_token")
	form.Set("redirect_uri", redirectURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, msTokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		data, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("refresh token request failed: %s: %s", res.Status, string(data))
	}

	var tr msTokenResponse
	if err := json.NewDecoder(res.Body).Decode(&tr); err != nil {
		return nil, err
	}

	return &tr, nil
}
