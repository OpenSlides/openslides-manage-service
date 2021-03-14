package clientutil

import "time"

// Config holds the top level arguments.
type Config struct {
	Address string
	Timeout time.Duration
}
