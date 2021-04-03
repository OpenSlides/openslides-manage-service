package datastore

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Get gets a fqField from the datastore.
func Get(ctx context.Context, addr string, key string, value interface{}) error {
	body, err := getRequestBody(key)
	if err != nil {
		return fmt.Errorf("building request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", addr+"/internal/datastore/reader/get", strings.NewReader(body))
	if err != nil {
		return fmt.Errorf("creating request to datastore: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("sending request to datastore: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			body = []byte("[can not read body]")
		}
		return fmt.Errorf("got response `%s`: %s", resp.Status, body)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %w", err)
	}

	var respData map[string]json.RawMessage
	if err := json.Unmarshal(respBody, &respData); err != nil {
		return fmt.Errorf("decoding response body `%s`: %w", respBody, err)
	}

	for field := range respData {
		if err := json.Unmarshal(respData[field], value); err != nil {
			return fmt.Errorf("decoding response field: %w", err)
		}
		return nil
	}
	return fmt.Errorf("datastore returned no data")
}

// Set sets a fqField at the datastore. Value has to be json.
func Set(ctx context.Context, key, value string) error {
	return nil
}

func getRequestBody(key string) (string, error) {
	parts := strings.Split(key, "/")
	if len(parts) != 3 {
		return "", fmt.Errorf("invalid key %s, expected two `/`", key)
	}
	return fmt.Sprintf(
		`{
			"fqid": "%s/%s",
			"mapped_fields": ["%s"]
		}`,
		parts[0], parts[1], parts[2],
	), nil
}
