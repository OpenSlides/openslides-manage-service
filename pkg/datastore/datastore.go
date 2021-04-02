package datastore

import "context"

// Get gets a fqField from the datastore.
func Get(ctx context.Context, key string) (string, error) {
	return "", nil
}

// Set sets a fqField at the datastore. Value has to be json.
func Set(ctx context.Context, key, value string) error {
	return nil
}
