package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const authHashPath = "/internal/auth/hash"

// Conn holds a connection to the auth service.
type Conn struct {
	authURL *url.URL
}

// New returns a new connection to the auth service.
func New(authURL *url.URL) *Conn {
	a := new(Conn)
	a.authURL = authURL
	return a
}

// Hash returns the hashed form of password as JSON.
func (a *Conn) Hash(ctx context.Context, password string) (string, error) {
	url := a.authURL
	url.Path = authHashPath
	reqBody := fmt.Sprintf(`{"toHash": "%s"}`, password)
	req, err := http.NewRequestWithContext(ctx, "POST", url.String(), strings.NewReader(reqBody))
	if err != nil {
		return "", fmt.Errorf("creating request to auth service: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("sending request: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			body = []byte("[can not read body]")
		}
		return "", fmt.Errorf("auth service returned %s: %s", resp.Status, body)
	}

	var respBody struct {
		Hash string `json:"hash"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return "", fmt.Errorf("decoding response: %w", err)
	}

	return respBody.Hash, nil
}
