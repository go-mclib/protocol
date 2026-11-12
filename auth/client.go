package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	ns "github.com/go-mclib/protocol/javaproto/wire"
	browser "github.com/pkg/browser"
)

// Client is the Microsoft/Xbox/Minecraft authentication client
type Client struct {
	cfg        Config
	httpClient *http.Client
	tokenStore TokenStore
	username   string
}

// Config is the configuration for the auth client
type Config struct {
	ClientID     string
	RedirectPort int
	Scopes       []string
	HTTPClient   *http.Client
	// TokenStore persists sessions. Nil uses the default file store
	// (~/.mclib/credentials_cache.json); if that cannot be created, sessions
	// are not cached.
	TokenStore TokenStore
	Username   string
}

// LoginData is the data returned from a login. It is also the cached session format.
type LoginData struct {
	AccessToken  string    `json:"accessToken"`
	RefreshToken string    `json:"refreshToken"`
	UUID         ns.UUID   `json:"uuid"`
	Username     string    `json:"username"`
	ExpiresAt    time.Time `json:"expiresAt"`
}

// UnmarshalJSON reads cached sessions. Caches migrated from the refresh-token-only
// format were written with an empty "uuid", which decodes to the nil UUID.
func (d *LoginData) UnmarshalJSON(data []byte) error {
	type plain LoginData
	var cached struct {
		plain
		UUID string `json:"uuid"`
	}
	if err := json.Unmarshal(data, &cached); err != nil {
		return err
	}
	decoded := LoginData(cached.plain)
	if cached.UUID != "" {
		uuid, err := ns.UUIDFromString(cached.UUID)
		if err != nil {
			return err
		}
		decoded.UUID = uuid
	}
	*d = decoded
	return nil
}

// IsValid returns true if the session has a valid access token that hasn't expired.
// Uses a 5-minute buffer to avoid edge cases.
func (d *LoginData) IsValid() bool {
	if d == nil || d.AccessToken == "" {
		return false
	}
	return time.Now().Add(5 * time.Minute).Before(d.ExpiresAt)
}

// NewClient creates a new Client with the given configuration
func NewClient(cfg Config) *Client {
	if cfg.RedirectPort == 0 {
		cfg.RedirectPort = freeLoopbackPort()
	}
	if len(cfg.Scopes) == 0 {
		cfg.Scopes = []string{"XboxLive.signin", "offline_access"}
	}

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 20 * time.Second}
	}

	store := cfg.TokenStore
	if store == nil {
		if s, err := NewFileTokenStore(""); err == nil {
			store = s
		}
	}

	return &Client{cfg: cfg, httpClient: httpClient, tokenStore: store, username: cfg.Username}
}

// AuthorizeWithLocalServer starts a local HTTP server and opens a browser to
// the Microsoft authorization URL. It returns the refresh token obtained by
// exchanging the authorization code. If the browser cannot be opened, the URL
// is printed to stderr for manual navigation.
func (c *Client) AuthorizeWithLocalServer(ctx context.Context) (string, error) {
	if c.cfg.ClientID == "" {
		return "", errors.New("missing client_id in Config")
	}

	state := newOAuthState()
	codeCh := make(chan string, 1)
	srv, port, err := startLocalServer(c.cfg.RedirectPort, state, codeCh)
	if err != nil {
		return "", err
	}
	defer stopLocalServer(srv)

	redirectURL := fmt.Sprintf("http://127.0.0.1:%d", port)
	authURL := buildAuthorizeURL(c.cfg.ClientID, redirectURL, state, c.cfg.Scopes)
	if err := browser.OpenURL(authURL); err != nil {
		// headless environments: keep the callback server up for manual navigation
		fmt.Fprintf(os.Stderr, "failed to open browser (%v); open this URL to sign in: %s\n", err, authURL)
	}

	// wait for code or cancellation
	select {
	case <-ctx.Done():
		return "", fmt.Errorf("authentication canceled: %w", ctx.Err())
	case code := <-codeCh:
		if code == "" {
			return "", errors.New("failed to obtain authorization code")
		}
		form := url.Values{"grant_type": {"authorization_code"}, "code": {code}}
		tokenRes, err := requestMSToken(ctx, c.httpClient, c.cfg.ClientID, redirectURL, form, "token exchange")
		if err != nil {
			return "", err
		}
		if tokenRes.RefreshToken == "" {
			return "", errors.New("no refresh_token in token response")
		}
		return tokenRes.RefreshToken, nil
	}
}

// LoginWithRefreshToken performs a login with a refresh token.
func (c *Client) LoginWithRefreshToken(ctx context.Context, refreshToken string) (LoginData, error) {
	if c.cfg.ClientID == "" {
		return LoginData{}, errors.New("missing client_id in Config")
	}
	if refreshToken == "" {
		return LoginData{}, errors.New("empty refresh token")
	}
	redirectURL := fmt.Sprintf("http://127.0.0.1:%d", c.cfg.RedirectPort)

	// refresh Microsoft access token
	form := url.Values{"grant_type": {"refresh_token"}, "refresh_token": {refreshToken}}
	tokRes, err := requestMSToken(ctx, c.httpClient, c.cfg.ClientID, redirectURL, form, "refresh token request")
	if err != nil {
		return LoginData{}, err
	}
	// the response may omit a rotated refresh token; keep the working one then
	if tokRes.RefreshToken != "" {
		refreshToken = tokRes.RefreshToken
	}

	// XBL authenticate
	xblRes, err := xblAuthenticate(ctx, c.httpClient, tokRes.AccessToken)
	if err != nil {
		return LoginData{}, err
	}
	if len(xblRes.DisplayClaims.XUI) == 0 {
		return LoginData{}, errors.New("xbl response has no user hash")
	}

	// XSTS authorize
	xstsRes, err := xstsAuthorize(ctx, c.httpClient, xblRes.Token)
	if err != nil {
		return LoginData{}, err
	}

	// Minecraft login with Xbox
	mcAuth, err := minecraftLoginWithXbox(ctx, c.httpClient, xblRes.DisplayClaims.XUI[0].UHS, xstsRes.Token)
	if err != nil {
		return LoginData{}, err
	}

	// verify entitlements
	owns, err := checkGameOwnership(ctx, c.httpClient, mcAuth.AccessToken)
	if err != nil {
		return LoginData{}, err
	}
	if !owns {
		return LoginData{}, errors.New("account does not own Minecraft (no entitlements)")
	}

	// fetch profile
	profile, err := fetchMinecraftProfile(ctx, c.httpClient, mcAuth.AccessToken)
	if err != nil {
		return LoginData{}, err
	}
	if profile == nil || profile.ID.IsNil() {
		return LoginData{}, errors.New("minecraft profile not found for account")
	}

	// calculate expiry time from expires_in (seconds)
	expiresAt := time.Now().Add(time.Duration(mcAuth.ExpiresIn) * time.Second)

	return LoginData{
		AccessToken:  mcAuth.AccessToken,
		RefreshToken: refreshToken,
		UUID:         profile.ID,
		Username:     profile.Name,
		ExpiresAt:    expiresAt,
	}, nil
}

// Login performs a cached login. It first checks for a valid cached session
// (with unexpired access token). If no valid session exists, it attempts to
// refresh using the stored refresh token. If that fails, it falls back to
// interactive auth via a local HTTP callback and browser.
//
// The username for caching is determined by:
// 1. The Username field in Config if set
// 2. The username from the LoginData response (after successful login)
func (c *Client) Login(ctx context.Context) (LoginData, error) {
	store := c.tokenStore

	// try cached session if username is specified
	if store != nil && c.username != "" {
		if session, err := store.LoadSession(c.username); err == nil && session != nil {
			// check if session is still valid (access token not expired)
			if session.IsValid() {
				c.username = session.Username
				return *session, nil
			}

			// session expired but we have a refresh token - try to refresh
			if session.RefreshToken != "" {
				if data, err := c.LoginWithRefreshToken(ctx, session.RefreshToken); err == nil {
					c.username = data.Username
					_ = store.SaveSession(&data) // caching is best-effort
					return data, nil
				}
			}
		}
	}

	// no cache or cache failed; must reauthenticate
	rt, err := c.AuthorizeWithLocalServer(ctx)
	if err != nil {
		return LoginData{}, err
	}
	data, err := c.LoginWithRefreshToken(ctx, rt)
	if err != nil {
		return LoginData{}, err
	}

	// save session with the username from login response
	c.username = data.Username
	if store != nil {
		_ = store.SaveSession(&data) // caching is best-effort
	}
	return data, nil
}

// Username returns the current username for this client.
func (c *Client) Username() string {
	return c.username
}

// freeLoopbackPort asks the OS for a free loopback port, or returns 0 so the
// callback server picks one when it starts.
func freeLoopbackPort() int {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0
	}
	defer ln.Close()
	return ln.Addr().(*net.TCPAddr).Port
}
