package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	ns "github.com/go-mclib/protocol/javaproto/wire"
)

const (
	mcAuthLoginWithXboxURL = "https://api.minecraftservices.com/authentication/login_with_xbox"
	mcEntitlementsURL      = "https://api.minecraftservices.com/entitlements/mcstore"
	mcProfileURL           = "https://api.minecraftservices.com/minecraft/profile"
)

type minecraftLoginRequest struct {
	IdentityToken string `json:"identityToken"`
}

type minecraftLoginResponse struct {
	Username    string `json:"username"`
	Roles       []any  `json:"roles"`
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

func minecraftLoginWithXbox(ctx context.Context, httpClient *http.Client, userHash, xstsToken string) (*minecraftLoginResponse, error) {
	body := minecraftLoginRequest{IdentityToken: fmt.Sprintf("XBL3.0 x=%s;%s", userHash, xstsToken)}
	var out minecraftLoginResponse
	if err := postJSON(ctx, httpClient, mcAuthLoginWithXboxURL, body, "minecraft login_with_xbox", &out); err != nil {
		return nil, err
	}
	return &out, nil
}

type entitlementsResponse struct {
	Items []struct {
		Name      string `json:"name"`
		Signature string `json:"signature"`
	} `json:"items"`
	Signature string `json:"signature"`
	KeyID     string `json:"keyId"`
}

func checkGameOwnership(ctx context.Context, httpClient *http.Client, mcAccessToken string) (bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, mcEntitlementsURL, nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("Authorization", "Bearer "+mcAccessToken)

	var out entitlementsResponse
	if err := doJSON(httpClient, req, "entitlements request", &out); err != nil {
		return false, err
	}

	if len(out.Items) == 0 {
		return false, nil
	}

	// according to wiki, expect product_minecraft and game_minecraft
	// https://minecraft.wiki/w/Microsoft_authentication#Checking_game_ownership
	var hasProduct, hasGame bool
	for _, it := range out.Items {
		switch it.Name {
		case "product_minecraft":
			hasProduct = true
		case "game_minecraft":
			hasGame = true
		}
	}

	return hasProduct && hasGame, nil
}

type profile struct {
	ID    ns.UUID `json:"id"`
	Name  string  `json:"name"`
	Skins any     `json:"skins"`
	Capes any     `json:"capes"`
}

func fetchMinecraftProfile(ctx context.Context, httpClient *http.Client, mcAccessToken string) (*profile, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, mcProfileURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+mcAccessToken)

	var out profile
	if err := doJSON(httpClient, req, "profile request", &out); err != nil {
		var statusErr *statusError
		if errors.As(err, &statusErr) && statusErr.code == http.StatusNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &out, nil
}
