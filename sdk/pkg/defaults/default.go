// Package defaults provides helpers for declaring default values of common
// non-primitive types (durations, timestamps, CIDR networks) in generated
// flag definitions.
package defaults

import (
	"net"
	"time"

	"github.com/gocloud9/gen-cobra-flags/sdk/pkg/it"
)

// ParseDuration parses a Go duration string and panics on error. It is meant
// for declaring constant defaults at package initialisation, where an invalid
// literal is a programming error.
func ParseDuration(im string) time.Duration {
	return it.Must(time.ParseDuration(im))
}

// ParseTime parses im using the given reference layout and panics on error. It
// is meant for declaring constant default timestamps.
func ParseTime(layout, im string) time.Time {
	return it.Must(time.Parse(layout, im))
}

// ParseCIDR parses a CIDR notation network and panics on error, returning the
// masked network. It is meant for declaring constant default networks.
func ParseCIDR(cidr string) *net.IPNet {
	_, cidrNet, err := net.ParseCIDR(cidr)
	if err != nil {
		panic(err)
	}

	return cidrNet
}
