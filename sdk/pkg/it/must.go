// Package it provides small generic helpers for working with the common
// (value, error) return convention, most notably Must, which panics on error.
package it

import "time"

// Must returns out, panicking if err is non-nil. It unwraps a (value, error)
// pair into a single value for use in variable initialisers where an error
// indicates a programming mistake.
func Must[T any](out T, err error) T {
	if err != nil {
		panic(err)
	}

	return out
}

// ParseDuration parses a Go duration string and panics on error.
func ParseDuration(im string) time.Duration {
	return Must(time.ParseDuration(im))
}
