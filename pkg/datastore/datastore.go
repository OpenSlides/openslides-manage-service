package datastore

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
	writeSubpath  = "/write"
)

// Conn holds a connection to the datastore service (reader and writer).
type Conn struct {
	ctx       context.Context
	readerURL *url.URL
	writerURL *url.URL
}

// New returns a new connection to the datastore.
func New(ctx context.Context, readerURL *url.URL, writerURL *url.URL) *Conn {
	d := new(Conn)
	d.ctx = ctx
	d.readerURL = readerURL
	d.writerURL = writerURL
	return d
}

// Exists does check if a collection object with given id exists.
func (d *Conn) Exists(collection string, id int) (bool, error) {
	reqBody := fmt.Sprintf(
		`{
			"collection": "%s",
			"filter": {
				"field": "id",
				"value": %d,
				"operator": "="
			}
		}`,
		collection, id,
	)
	addr := d.readerURL.String() + existsSubpath

	req, err := http.NewRequestWithContext(d.ctx, "POST", addr, strings.NewReader(reqBody))
	if err != nil {
		return false, fmt.Errorf("creating request to datastore: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("sending request to datastore at %s: %w", addr, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			body = []byte("[can not read body]")
		}
		return false, fmt.Errorf("got response `%s`: %s", resp.Status, body)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("reading response body: %w", err)
	}

	var respData struct {
		Exists bool `json:"exists"`
	}
	if err := json.Unmarshal(respBody, &respData); err != nil {
		return false, fmt.Errorf("decoding response body `%s`: %w", respBody, err)
	}
	return respData.Exists, nil
}

// Create sends a create event to the datastore.
func (d *Conn) Create(fqid string, fields map[string]json.RawMessage) error {
	f, err := json.Marshal(fields)
	if err != nil {
		return fmt.Errorf("marshalling fields: %w", err)
	}

	reqBody := fmt.Sprintf(
		`{
			"user_id": 0,
			"information": {},
			"locked_fields":{}, "events":[
				{"type":"create","fqid":"%s","fields":%s}
			]
		}`,
		fqid, string(f),
	)

	if err := sendWriteRequest(d.ctx, d.writerURL, reqBody); err != nil {
		return fmt.Errorf("sending write request to datastore: %w", err)
	}

	return nil
}

// sendWriteRequest sends the give request body to the datastore.
func sendWriteRequest(ctx context.Context, writerURL *url.URL, reqBody string) error {
	addr := writerURL.String() + writeSubpath

	req, err := http.NewRequestWithContext(ctx, "POST", addr, strings.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("creating request to datastore: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("sending request to datastore at %s: %w", addr, err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			body = []byte("[can not read body]")
		}
		return fmt.Errorf("got response `%s`: %s", resp.Status, body)
	}

	return nil
}
