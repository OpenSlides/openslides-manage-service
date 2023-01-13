package backendaction

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
	route                string
}

const (
	// ActionRoute is used to mark the connection to be usable for a route to an
	// action (called action_route or internal_action_route in the backend).
	ActionRoute = "action"

	// MigrationsRoute is used to mark the connection to be usable for a route
	// to the migrations handler.
	MigrationsRoute = "migrations"

	// HealthRoute is used to mark the connection to be usable for a route to
	// the backend with a health request.
	HealthRoute = "health"
)

// New returns a new connection to the backend action service.
func New(url *url.URL, pw []byte, route string) *Conn {
	c := new(Conn)
	c.URL = url
	c.internalAuthPassword = pw
	c.route = route
	return c
}

// Single sends a request to backend action service with a single action.
func (c *Conn) Single(ctx context.Context, name string, data json.RawMessage) (json.RawMessage, error) {
	if c.route != ActionRoute {
		return nil, fmt.Errorf("invalid route for this connection; expected %q, got %q", ActionRoute, c.route)
	}

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

	res, err := requestWithPassword(ctx, "POST", c.URL.String(), c.internalAuthPassword, bytes.NewReader(encodedBody))
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}
	// Hint: res is something like
	// {"success": ..., "message": ..., "results": [[{"id": 42}, {"id": 42}]]}
	// depending on the respective action

	var content struct {
		Success bool              `json:"success"`
		Message string            `json:"message"`
		Results []json.RawMessage // We deconstruct only the outer list and forward the inner list to the caller.
	}
	if err := json.Unmarshal(res, &content); err != nil {
		return nil, fmt.Errorf("unmarshalling response body: %w", err)
	}
	if len(content.Results) != 1 {
		return nil, fmt.Errorf("response body content should have one item, but has %d", len(content.Results))
	}

	return content.Results[0], nil
}

// Migrations sends the given migrations command to the backend.
func (c *Conn) Migrations(ctx context.Context, command string) (json.RawMessage, error) {
	if c.route != MigrationsRoute {
		return nil, fmt.Errorf("invalid route for this connection; expected %q, got %q", MigrationsRoute, c.route)
	}

	reqBody := struct {
		Cmd string `json:"cmd"`
	}{
		Cmd: command,
	}
	encodedBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshalling request body: %w", err)
	}

	res, err := requestWithPassword(ctx, "POST", c.URL.String(), c.internalAuthPassword, bytes.NewReader(encodedBody))
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}
	return res, nil
}

// Health sends the health request to the backend.
func (c *Conn) Health(ctx context.Context) (json.RawMessage, error) {
	if c.route != HealthRoute {
		return nil, fmt.Errorf("invalid route for this connection; expected %q, got %q", HealthRoute, c.route)
	}

	res, err := requestWithPassword(ctx, "GET", c.URL.String(), c.internalAuthPassword, nil)
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}
	return res, nil
}

func requestWithPassword(ctx context.Context, method string, addr string, pw []byte, body io.Reader) (json.RawMessage, error) {
	req, err := http.NewRequestWithContext(ctx, method, addr, body)
	if err != nil {
		return nil, fmt.Errorf("creating request to backend action service: %w", err)
	}
	creds := shared.BasicAuth{
		Password: pw,
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(shared.AuthHeader, creds.EncPassword())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending request to backend action service at %q: %w", addr, err)
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

	if !json.Valid(encodedResp) {
		return nil, fmt.Errorf("response body does not contain valid JSON, got %q", string(encodedResp))
	}

	return json.RawMessage(encodedResp), nil
}
