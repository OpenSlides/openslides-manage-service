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
	"github.com/rs/zerolog/log"
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
	const maxRetries = 5
	const backoffDelay = 4 * time.Second
	const requestTimeout = 5 * time.Second // Individual request timeout

	var originalBody []byte
	if body != nil {
		var err error
		originalBody, err = io.ReadAll(body)
		if err != nil {
			return nil, fmt.Errorf("reading request body: %w", err)
		}
	}

	// Create a custom HTTP client that doesn't have its own timeout
	httpClient := &http.Client{
		Timeout: 0, // No client-level timeout, let request context handle it
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second, // This is the dial timeout
				KeepAlive: 30 * time.Second,
			}).DialContext,
			ResponseHeaderTimeout: 0, // Let request context handle it
			IdleConnTimeout:       90 * time.Second,
		},
	}

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		// Check if parent context is already done before starting attempt
		if ctx.Err() != nil {
			log.Warn().Msg("parent context done before retry attempt")
			return nil, ctx.Err()
		}

		var bodyReader io.Reader
		if originalBody != nil {
			bodyReader = bytes.NewReader(originalBody)
		}

		// Create a fresh context for this individual request
		// This gives each request its own timeout window
		reqCtx, cancel := context.WithTimeout(context.Background(), requestTimeout)

		req, err := http.NewRequestWithContext(reqCtx, method, addr, bodyReader)
		if err != nil {
			cancel()
			return nil, fmt.Errorf("creating request: %w", err)
		}

		creds := shared.BasicAuth{Password: pw}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set(shared.AuthHeader, creds.EncPassword())

		log.Warn().Int("attempt", attempt+1).Int("max_retries", maxRetries).Msg("attempting request")

		resp, err := httpClient.Do(req)
		cancel() // Always clean up the request context

		if err != nil {
			if isNetworkError(err) {
				lastErr = err
				log.Warn().Err(err).Msgf("network error on attempt %d/%d; retrying...", attempt+1, maxRetries)

				// Only retry if we haven't exhausted all attempts
				if attempt < maxRetries-1 {
					log.Warn().Dur("backoff_delay", backoffDelay).Msg("waiting before retry")

					// Wait for backoff delay or until parent context is done
					select {
					case <-time.After(backoffDelay):
						log.Warn().Msg("backoff complete, retrying")
						continue
					case <-ctx.Done():
						log.Warn().Msg("parent context done during backoff")
						return nil, ctx.Err()
					}
				}
				// If we're here, we've exhausted all retries
				log.Warn().Err(lastErr).Msgf("request failed after %d attempts", maxRetries)
				return nil, fmt.Errorf("request failed after %d attempts: %w", maxRetries, lastErr)
			}
			// Non-network error - don't retry
			return nil, fmt.Errorf("non-network error: %w", err)
		}

		// Success - process the response
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

		log.Warn().Msg("request succeeded")
		return json.RawMessage(encodedResp), nil
	}

	// This should never be reached due to the logic above, but keeping it for safety
	return nil, fmt.Errorf("request failed after %d attempts: %w", maxRetries, lastErr)
}

func isNetworkError(err error) bool {
	// Unwrap *url.Error first
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		err = urlErr.Err
	}

	// Check for context deadline exceeded (should be treated as network error for retry)
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	// Check for context canceled (should NOT be retried)
	if errors.Is(err, context.Canceled) {
		return false
	}

	// Timeout or temporary network error
	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}

	// Lower-level operation error (e.g., connection refused)
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		return true
	}

	// Unexpected connection closures
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return true
	}

	// Check for specific network-related error strings
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
