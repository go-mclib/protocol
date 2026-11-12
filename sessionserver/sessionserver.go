package sessionserver

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/go-mclib/protocol/crypto"
	ns "github.com/go-mclib/protocol/javaproto/wire"
)

const userAgent = "gomc-lib/protocol (github.com/go-mclib/protocol)"

// defaultBaseURL is Mojang's session server.
const defaultBaseURL = "https://sessionserver.mojang.com"

// Client calls the session server's join and hasJoined endpoints.
type Client struct {
	// BaseURL is the session server root, without a trailing slash.
	BaseURL    string
	httpClient *http.Client
}

// New returns a client for Mojang's session server that sends
// requests through httpClient; nil uses a client with a 10 s timeout.
func New(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	return &Client{BaseURL: defaultBaseURL, httpClient: httpClient}
}

// joinRequest is the request payload for /session/minecraft/join
type joinRequest struct {
	AccessToken     string `json:"accessToken"`
	SelectedProfile string `json:"selectedProfile"`
	ServerID        string `json:"serverId"`
}

// HasJoinedResponse represents the response from /session/minecraft/hasJoined
type HasJoinedResponse struct {
	ID         ns.UUID    `json:"id"`
	Name       string     `json:"name"`
	Properties []Property `json:"properties"`
}

// Property represents a profile property
type Property struct {
	Name      string `json:"name"`
	Value     string `json:"value"`
	Signature string `json:"signature,omitempty"`
}

// errorResponse is an error response from Mojang
type errorResponse struct {
	Error        string `json:"error"`
	ErrorMessage string `json:"errorMessage,omitempty"`
	Path         string `json:"path,omitempty"`
}

func (e errorResponse) String() string {
	if e.ErrorMessage != "" {
		return fmt.Sprintf("%s: %s (path: %s)", e.Error, e.ErrorMessage, e.Path)
	}
	return fmt.Sprintf("%s (path: %s)", e.Error, e.Path)
}

// Join registers the client's intent to join the server identified by
// serverID, sharedSecret and the server's DER public key.
func (c *Client) Join(ctx context.Context, accessToken string, selectedProfile ns.UUID, serverID string, sharedSecret, publicKey []byte) error {
	if !validateAccessToken(accessToken) {
		return errors.New("invalid access token format")
	}

	jsonData, err := json.Marshal(joinRequest{
		AccessToken:     accessToken,
		SelectedProfile: hex.EncodeToString(selectedProfile[:]),
		ServerID:        computeServerHash(serverID, sharedSecret, publicKey),
	})
	if err != nil {
		return fmt.Errorf("failed to marshal join request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/session/minecraft/join", bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create join request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	status, body, err := c.do(req)
	if err != nil {
		return fmt.Errorf("join request: %w", err)
	}
	if status == http.StatusNoContent {
		return nil
	}
	return apiError("authentication", status, body)
}

// HasJoined reports the profile of username if it joined the server identified
// by serverID, sharedSecret and the server's DER public key, or nil if it did
// not. A non-empty ip additionally requires the join to come from that address.
func (c *Client) HasJoined(ctx context.Context, username, serverID string, sharedSecret, publicKey []byte, ip string) (*HasJoinedResponse, error) {
	query := url.Values{"username": {username}, "serverId": {computeServerHash(serverID, sharedSecret, publicKey)}}
	if ip != "" {
		query.Set("ip", ip)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+"/session/minecraft/hasJoined?"+query.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create hasJoined request: %w", err)
	}

	status, body, err := c.do(req)
	if err != nil {
		return nil, fmt.Errorf("hasJoined request: %w", err)
	}
	switch status {
	case http.StatusNoContent:
		return nil, nil
	case http.StatusOK:
	default:
		return nil, apiError("hasJoined", status, body)
	}

	var hasJoinedResp HasJoinedResponse
	if err := json.Unmarshal(body, &hasJoinedResp); err != nil {
		return nil, fmt.Errorf("failed to parse hasJoined response: %w", err)
	}
	return &hasJoinedResp, nil
}

// do sends req and returns the response status and body.
func (c *Client) do(req *http.Request) (int, []byte, error) {
	req.Header.Set("User-Agent", userAgent)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to read response body: %w", err)
	}
	return resp.StatusCode, body, nil
}

// apiError describes a failed call, using Mojang's error payload when present.
func apiError(op string, status int, body []byte) error {
	var errResp errorResponse
	if err := json.Unmarshal(body, &errResp); err != nil {
		return fmt.Errorf("%s failed: %s (status %d)", op, body, status)
	}
	return fmt.Errorf("%s failed: %s (status %d)", op, errResp, status)
}

// computeServerHash returns the Minecraft hex digest of serverID, the shared
// secret and the server's public key, as sent to the session server.
func computeServerHash(serverID string, sharedSecret, publicKey []byte) string {
	return crypto.MinecraftHexDigest([]byte(serverID), sharedSecret, publicKey)
}

func validateAccessToken(token string) bool {
	return len(token) > 10 && len(token) < 2048
}
