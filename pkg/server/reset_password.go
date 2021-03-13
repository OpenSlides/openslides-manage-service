package server

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	hashPath  = "/internal/auth/hash"
	writePath = "/internal/datastore/writer/write"
)

// hashPassword returns the hashed form of password as a json-value.
// json-value means, that the string already contains the necessary "".
func hashPassword(ctx context.Context, cfg *Config, password string) (string, error) {
	reqBody := fmt.Sprintf(`{"toHash": "%s"}`, password)
	req, err := http.NewRequestWithContext(ctx, "POST", cfg.AuthAddr()+hashPath, strings.NewReader(reqBody))
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

	hash, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading body: %w", err)
	}
	return string(hash), nil
}

func setPassword(ctx context.Context, cfg *Config, userID int, hash string) error {
	reqBody := fmt.Sprintf(`{"user_id":1,"information":{},"locked_fields":{},"events":[{"type":"update","fqid":"user/%d","fields":{"password":%s}}]}`, userID, hash)
	req, err := http.NewRequestWithContext(ctx, "POST", cfg.DSWriterAddr()+writePath, strings.NewReader(reqBody))
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
		return fmt.Errorf("datastore service returned %s: %s", resp.Status, body)
	}

	return nil
}
