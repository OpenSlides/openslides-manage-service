package backendaction

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

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

func executeRequest(ctx context.Context, method string, addr string, pw []byte, bodyReader io.Reader, requestTimeout time.Duration) (json.RawMessage, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	reqCtx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, method, addr, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	creds := shared.BasicAuth{Password: pw}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(shared.AuthHeader, creds.EncPassword())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server error response %d: %s", resp.StatusCode, string(respBody))
	}

	encodedResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	if !json.Valid(encodedResp) {
		return nil, fmt.Errorf("invalid JSON response: %q", string(encodedResp))
	}

	return json.RawMessage(encodedResp), nil
}

func requestWithPassword(ctx context.Context, method string, addr string, pw []byte, body io.Reader) (json.RawMessage, error) {
	const maxRetries = 5
	const backoffDelay = 4 * time.Second
	const requestTimeout = 5 * time.Second

	var originalBody []byte
	if body != nil {
		var err error
		originalBody, err = io.ReadAll(body)
		if err != nil {
			return nil, fmt.Errorf("reading request body: %w", err)
		}
	}

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		var bodyReader io.Reader
		if originalBody != nil {
			bodyReader = bytes.NewReader(originalBody)
		}

		result, err := executeRequest(ctx, method, addr, pw, bodyReader, requestTimeout)
		if err == nil {
			return result, nil
		}

		lastErr = err
		if isNetworkError(err) && attempt < maxRetries-1 {
			select {
			case <-time.After(backoffDelay):
				continue
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
		break
	}

	return nil, fmt.Errorf("request failed after %d attempts: %w", maxRetries, lastErr)
}

func isNetworkError(err error) bool {
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		err = urlErr.Err
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	if errors.Is(err, context.Canceled) {
		return false
	}
	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		return true
	}
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return true
	}
	errStr := err.Error()
	networkErrors := []string{
		"connection refused",
		"connection reset by peer",
		"i/o timeout",
		"no such host",
		"network is unreachable",
		"connection timed out",
	}
	for _, netErrStr := range networkErrors {
		if strings.Contains(strings.ToLower(errStr), netErrStr) {
			return true
		}
	}
	return false
}
