package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// statusError reports a non-2xx response from an auth endpoint.
type statusError struct {
	op     string
	status string
	code   int
	body   []byte
}

func (e *statusError) Error() string {
	return fmt.Sprintf("%s failed: %s: %s", e.op, e.status, e.body)
}

// doJSON sends req and decodes a 2xx JSON response into out.
func doJSON(httpClient *http.Client, req *http.Request, op string, out any) error {
	res, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		body, _ := io.ReadAll(res.Body) // best-effort detail for the error
		return &statusError{op: op, status: res.Status, code: res.StatusCode, body: body}
	}
	if err := json.NewDecoder(res.Body).Decode(out); err != nil {
		return fmt.Errorf("%s: decode response: %w", op, err)
	}
	return nil
}

// postJSON POSTs in as JSON to url and decodes the JSON response into out.
func postJSON(ctx context.Context, httpClient *http.Client, url string, in any, op string, out any) error {
	body, err := json.Marshal(in)
	if err != nil {
		return fmt.Errorf("%s: encode request: %w", op, err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	return doJSON(httpClient, req, op, out)
}
