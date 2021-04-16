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
	getSubpath   = "/get"
	writeSubpath = "/write"
)

// Get fetches a FQField from the datastore.
func Get(ctx context.Context, readerURL *url.URL, fqfield string, value interface{}) error {
	keyParts := strings.Split(fqfield, "/")
	if len(keyParts) != 3 {
		return fmt.Errorf("invalid FQField %s, expected two `/`", fqfield)
	}

	reqBody := fmt.Sprintf(
		`{
			"fqid": "%s/%s",
			"mapped_fields": ["%s"]
		}`,
		keyParts[0], keyParts[1], keyParts[2],
	)
	addr := readerURL.String() + getSubpath

	req, err := http.NewRequestWithContext(ctx, "POST", addr, strings.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("creating request to datastore: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("sending request to datastore at %s: %w", addr, err)
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

	if err := json.Unmarshal(respData[keyParts[2]], value); err != nil {
		return fmt.Errorf("decoding response field: %w", err)
	}

	return nil
}

// Set sets a FQField at the datastore. Value has to be JSON.
func Set(ctx context.Context, writerURL *url.URL, fqfield string, value json.RawMessage) error {
	parts := strings.Split(fqfield, "/")
	if len(parts) != 3 {
		return fmt.Errorf("invalid FQField %s, expected two `/`", fqfield)
	}

	reqBody := fmt.Sprintf(
		`{
			"user_id": 0,
			"information": {},
			"locked_fields":{}, "events":[
				{"type":"update","fqid":"%s/%s","fields":{"%s":%s}}
			]
		}`,
		parts[0], parts[1], parts[2], value,
	)

	if err := sendWriteRequest(ctx, writerURL, reqBody); err != nil {
		return fmt.Errorf("sending write request to datastore: %w", err)
	}

	return nil
}

// Create sends a create event to the datastore.
func Create(ctx context.Context, writerURL *url.URL, fqid string, fields map[string]json.RawMessage) error {
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

	if err := sendWriteRequest(ctx, writerURL, reqBody); err != nil {
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

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			body = []byte("[can not read body]")
		}
		return fmt.Errorf("got response `%s`: %s", resp.Status, body)
	}

	return nil
}
