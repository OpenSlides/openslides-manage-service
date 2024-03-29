package datastorereader

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const (
	existsSubpath = "/exists"
	getAllSubpath = "/get_all"
	filterSubpath = "/filter"
)

// Conn holds a connection to the datastoreReader service.
type Conn struct {
	readerURL *url.URL
}

// New returns a new connection to the datastore.
func New(readerURL *url.URL) *Conn {
	d := new(Conn)
	d.readerURL = readerURL
	return d
}

// Exists does check if a collection object with given id exists.
func (d *Conn) Exists(ctx context.Context, collection string, filter string) (bool, error) {
	reqBody := makeRequestBody(collection, filter, "")
	addr := d.readerURL.String() + existsSubpath

	respBody, err := sendReadRequest(ctx, addr, reqBody)
	if err != nil {
		return false, fmt.Errorf("initiating datastore read request: %w", err)
	}

	var respData struct {
		Exists bool `json:"exists"`
	}
	if err := json.Unmarshal(respBody, &respData); err != nil {
		return false, fmt.Errorf("decoding response body `%s`: %w", respBody, err)
	}
	return respData.Exists, nil
}

// Filter searches for the fitting model and also restricts to fields if provided
func (d *Conn) Filter(ctx context.Context, collection string, filter string, fields string) (string, error) {
	reqBody := makeRequestBody(collection, filter, fields)
	addr := d.readerURL.String() + filterSubpath

	respBody, err := sendReadRequest(ctx, addr, reqBody)
	if err != nil {
		return "", fmt.Errorf("initiating datastore read request: %w", err)
	}

	var respData struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(respBody, &respData); err != nil {
		return "", fmt.Errorf("decoding response body `%s`: %w", respBody, err)
	}
	return string(respData.Data[:]), nil
}

// GetAll gets all models in the given collection as json object
func (d *Conn) GetAll(ctx context.Context, collection string, fields string) (string, error) {
	reqBody := makeRequestBody(collection, "", fields)
	addr := d.readerURL.String() + getAllSubpath

	respBody, err := sendReadRequest(ctx, addr, reqBody)
	if err != nil {
		return "", fmt.Errorf("initiating datastore read request: %w", err)
	}

	var respData json.RawMessage
	if err := json.Unmarshal(respBody, &respData); err != nil {
		return "", fmt.Errorf("decoding response body `%s`: %w", respBody, err)
	}
	return string(respData[:]), nil
}

// sendReadRequest sends the given request body to the datastore.
func sendReadRequest(ctx context.Context, addr string, reqBody string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", addr, strings.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("creating request to datastore: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending request to datastore at %s: %w", addr, err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			body = []byte("[can not read body]")
		}
		return nil, fmt.Errorf("got response `%s`: %s", resp.Status, body)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	return respBody, nil
}

func makeRequestBody(collection string, filter string, fields string) string {
	body := fmt.Sprintf(`"collection": "%s"`, collection)
	if filter != "" {
		body += fmt.Sprintf(`, "filter": %s`, filter)
	}
	if fields != "" {
		body += fmt.Sprintf(`, "mapped_fields": %s`, fields)
	}
	return fmt.Sprintf(`{ %s }`, body)
}
