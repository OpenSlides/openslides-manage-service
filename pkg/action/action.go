package action

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/OpenSlides/openslides-manage-service/pkg/shared"
)

// Conn holds a connection to the backend action service.
type Conn struct {
	URL                  *url.URL
	internalAuthPassword []byte
}

// New returns a new connection to the backend action service.
func New(url *url.URL, pw []byte) *Conn {
	c := new(Conn)
	c.URL = url
	c.internalAuthPassword = pw
	return c
}

// Single sends a request to backend action service with a single action.
func (c *Conn) Single(ctx context.Context, name string, data json.RawMessage) (json.RawMessage, error) {
	addr := c.URL.String()
	reqBody := []struct {
		Action string          `json:"action"`
		Data   json.RawMessage `json:"data"`
	}{
		{
			Action: name,
			Data:   data,
		},
	}
	encodedBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshalling request body: %w", err)
	}
	// Hint: encodedBody is something like
	// [{"action": "user.create", "data": [{"username": "foo", ...}, {"username": "bar", ...}]}]
	req, err := http.NewRequestWithContext(ctx, "POST", addr, bytes.NewReader(encodedBody))
	if err != nil {
		return nil, fmt.Errorf("creating request to backend action service: %w", err)
	}
	creds := shared.BasicAuth{
		Password: c.internalAuthPassword,
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(shared.AuthHeader, creds.EncPassword())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending request to backend action service at %s: %w", addr, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			body = []byte("[can not read body]")
		}
		return nil, fmt.Errorf("got response %q: %q", resp.Status, body)
	}

	encodedResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}
	// Hint: encodedResp is something like
	// {"success": ..., "message": ..., "results": [[{"id": 42}, {"id": 42}]]}
	// depending on the respective action
	var content struct {
		Success bool              `json:"success"`
		Message string            `json:"message"`
		Results []json.RawMessage // We deconstruct only the outer list and forward the inner list to the caller.
	}
	if err := json.Unmarshal(encodedResp, &content); err != nil {
		return nil, fmt.Errorf("unmarshalling response body: %w", err)
	}
	if len(content.Results) != 1 {
		return nil, fmt.Errorf("response body content should have one item, but has %d", len(content.Results))
	}

	return content.Results[0], nil

}
