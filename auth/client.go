package auth

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"time"
)

// AuthClient is the main struct for the auth client
type AuthClient struct {
	cfg        AuthClientConfig
	httpClient *http.Client
	tokenStore TokenStore
	username   string
}

// AuthClientConfig is the configuration for the auth client
type AuthClientConfig struct {
	ClientID         string
	RedirectPort     int
	Scopes           []string
	HTTPClient       *http.Client
	TokenStore       TokenStore
	TokenStoreConfig TokenStoreConfig
	// Username is the Minecraft username to use for caching.
	// If empty, username from login response will be used.
	Username string
}

// LoginData is the data returned from a login
type LoginData struct {
	AccessToken  string
	RefreshToken string
	UUID         string
	Username     string
}

// NewClient creates a new AuthClient with the given configuration
func NewClient(cfg AuthClientConfig) *AuthClient {
	if cfg.RedirectPort == 0 {
		cfg.RedirectPort = tryPort()
	}
	if len(cfg.Scopes) == 0 {
		cfg.Scopes = []string{"XboxLive.signin", "offline_access"}
	}

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 20 * time.Second}
	}

	var store TokenStore
	if cfg.TokenStore != nil {
		store = cfg.TokenStore
	} else {
		// try to create store from config
		if s, err := NewTokenStore(cfg.TokenStoreConfig); err == nil {
			store = s
		}
	}

	return &AuthClient{cfg: cfg, httpClient: httpClient, tokenStore: store, username: cfg.Username}
}

// AuthorizeWithLocalServer starts a local HTTP server and opens a browser to
// the Microsoft authorization URL. It returns the authorization code, which
// can be used to exchange for tokens.
func (c *AuthClient) AuthorizeWithLocalServer(ctx context.Context) (string, error) {
	if c.cfg.ClientID == "" {
		return "", errors.New("missing client_id in Config")
	}

	redirectURL := fmt.Sprintf("http://127.0.0.1:%d", c.cfg.RedirectPort)
	authURL := buildAuthorizeURL(c.cfg.ClientID, redirectURL, c.cfg.Scopes)

	codeCh := make(chan string, 1)
	srv, err := startLocalServer(c.cfg.RedirectPort, codeCh)
	if err != nil {
		return "", err
	}
	defer func() { _ = stopLocalServer(srv) }()

	if err := openBrowser(authURL); err != nil {
		// Note: opening the browser can fail in headless environments; still allow manual navigation.
		// return the URL in the error for convenience.
		return "", fmt.Errorf("failed to open browser, navigate to: %s, err: %w", authURL, err)
	}

	// wait for code or cancellation
	select {
	case <-ctx.Done():
		return "", errors.New("authentication canceled")
	case code := <-codeCh:
		if code == "" {
			return "", errors.New("failed to obtain authorization code")
		}
		tokenRes, err := exchangeAuthCodeForTokens(ctx, c.httpClient, c.cfg.ClientID, redirectURL, code)
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
func (c *AuthClient) LoginWithRefreshToken(ctx context.Context, refreshToken string) (LoginData, error) {
	if c.cfg.ClientID == "" {
		return LoginData{}, errors.New("missing client_id in Config")
	}
	redirectURL := fmt.Sprintf("http://127.0.0.1:%d", c.cfg.RedirectPort)

	// refresh Microsoft access token
	tokRes, err := refreshAccessToken(ctx, c.httpClient, c.cfg.ClientID, redirectURL, refreshToken)
	if err != nil {
		return LoginData{}, err
	}
	msAccessToken := tokRes.AccessToken
	refreshToken = tokRes.RefreshToken

	// XBL authenticate
	xblRes, err := xblAuthenticate(ctx, c.httpClient, msAccessToken)
	if err != nil {
		return LoginData{}, err
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
	if profile == nil || profile.ID == "" {
		return LoginData{}, errors.New("minecraft profile not found for account")
	}

	return LoginData{
		AccessToken:  mcAuth.AccessToken,
		RefreshToken: refreshToken,
		UUID:         profile.ID,
		Username:     profile.Name,
	}, nil
}

// Login performs a cached login. It attempts to load a refresh token from the
// configured TokenStore (or the default file-based store), refreshes it, and
// completes the Microsoft/XBL/XSTS/Minecraft authentication flow. If no cached
// token exists or the refresh fails, it falls back to interactive auth via a
// local HTTP callback and browser, then saves the new refresh token.
//
// The username for caching is determined by:
// 1. The Username field in AuthClientConfig if set
// 2. The username from the LoginData response (after successful login)
func (c *AuthClient) Login(ctx context.Context) (LoginData, error) {
	store := c.tokenStore

	// try cached refresh token if username is specified
	if store != nil && c.username != "" {
		if rt, err := store.Load(c.username); err == nil && rt != "" {
			if data, err := c.LoginWithRefreshToken(ctx, rt); err == nil {
				// Update username in case it changed
				c.username = data.Username
				_ = store.Save(data.Username, data.RefreshToken)
				return data, nil
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

	// Save with the username from login response
	c.username = data.Username
	if store != nil {
		_ = store.Save(data.Username, data.RefreshToken)
	}

	return data, nil
}

// ClearCachedToken removes the stored refresh token for the current username.
// If username is not set, this returns an error.
func (c *AuthClient) ClearCachedToken(_ context.Context) error {
	if c.tokenStore == nil {
		return nil
	}

	if c.username == "" {
		return errors.New("no username set, cannot clear cached token")
	}

	return c.tokenStore.Clear(c.username)
}

// SetUsername updates the username for this client. This affects which cached
// token is loaded/saved.
func (c *AuthClient) SetUsername(username string) {
	c.username = username
}

// GetUsername returns the current username for this client.
func (c *AuthClient) GetUsername() string {
	return c.username
}

// ListCachedAccounts returns a list of all usernames that have cached tokens.
func (c *AuthClient) ListCachedAccounts() ([]string, error) {
	if c.tokenStore == nil {
		return nil, nil
	}
	return c.tokenStore.ListAccounts()
}

// tryPort tries to find an open port
func tryPort() int {
	randomPort := rand.Intn(65535-1024) + 1024

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", randomPort))
	if err != nil {
		return tryPort()
	}
	defer listener.Close()

	return randomPort
}
