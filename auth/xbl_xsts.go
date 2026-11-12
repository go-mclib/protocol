package auth

import (
	"context"
	"net/http"
)

const (
	xblUserAuthenticateURL = "https://user.auth.xboxlive.com/user/authenticate"
	xstsAuthorizeURL       = "https://xsts.auth.xboxlive.com/xsts/authorize"
)

type xblRequest struct {
	Properties struct {
		AuthMethod string `json:"AuthMethod"`
		SiteName   string `json:"SiteName"`
		RpsTicket  string `json:"RpsTicket"`
	} `json:"Properties"`
	RelyingParty string `json:"RelyingParty"`
	TokenType    string `json:"TokenType"`
}

// xboxTokenResponse is returned by both the XBL and XSTS endpoints.
type xboxTokenResponse struct {
	Token         string `json:"Token"`
	DisplayClaims struct {
		XUI []struct {
			UHS string `json:"uhs"`
		} `json:"xui"`
	} `json:"DisplayClaims"`
}

func xblAuthenticate(ctx context.Context, httpClient *http.Client, msAccessToken string) (*xboxTokenResponse, error) {
	body := xblRequest{
		RelyingParty: "http://auth.xboxlive.com",
		TokenType:    "JWT",
	}
	body.Properties.AuthMethod = "RPS"
	body.Properties.SiteName = "user.auth.xboxlive.com"
	body.Properties.RpsTicket = "d=" + msAccessToken

	var out xboxTokenResponse
	if err := postJSON(ctx, httpClient, xblUserAuthenticateURL, body, "xbl authenticate", &out); err != nil {
		return nil, err
	}
	return &out, nil
}

type xstsRequest struct {
	Properties struct {
		SandboxID  string   `json:"SandboxId"`
		UserTokens []string `json:"UserTokens"`
	} `json:"Properties"`
	RelyingParty string `json:"RelyingParty"`
	TokenType    string `json:"TokenType"`
}

func xstsAuthorize(ctx context.Context, httpClient *http.Client, xblToken string) (*xboxTokenResponse, error) {
	body := xstsRequest{
		RelyingParty: "rp://api.minecraftservices.com/",
		TokenType:    "JWT",
	}
	body.Properties.SandboxID = "RETAIL"
	body.Properties.UserTokens = []string{xblToken}

	var out xboxTokenResponse
	if err := postJSON(ctx, httpClient, xstsAuthorizeURL, body, "xsts authorize", &out); err != nil {
		return nil, err
	}
	return &out, nil
}
