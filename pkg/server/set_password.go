package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	authHashPath      = "/internal/auth/hash"
	datastorWritePath = "/internal/datastore/writer/write"
)

// hashPassword returns the hashed form of password as a JSON.
func hashPassword(ctx context.Context, cfg *Config, password string) (string, error) {
	reqBody := fmt.Sprintf(`{"toHash": "%s"}`, password)
	reqURL := cfg.AuthURL()
	reqURL.Path = authHashPath
	req, err := http.NewRequestWithContext(ctx, "POST", reqURL.String(), strings.NewReader(reqBody))
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

func setPassword(ctx context.Context, cfg *Config, userID int, hash string) error {
	reqBody := fmt.Sprintf(`{"user_id":0,"information":{},"locked_fields":{},"events":[{"type":"update","fqid":"user/%d","fields":{"password":"%s"}}]}`, userID, hash)
	reqURL := cfg.DatastoreWriterURL()
	reqURL.Path = datastorWritePath
	req, err := http.NewRequestWithContext(ctx, "POST", reqURL.String(), strings.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("sending request: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			body = []byte("[can not read body]")
		}
		return fmt.Errorf("datastore writer service returned %s: %s", resp.Status, body)
	}

	return nil
}
