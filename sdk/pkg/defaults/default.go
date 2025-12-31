package defaults

import (
	"github.com/gocloud9/gen-cobra-flags/sdk/pkg/it"
	"net"
	"time"
)

func ParseDuration(im string) time.Duration {
	return it.Must(time.ParseDuration(im))
}

func ParseTime(layout, im string) time.Time {
	return it.Must(time.Parse(layout, im))
}

func ParseCIDR(cidr string) *net.IPNet {
	_, cidrNet, err := net.ParseCIDR(cidr)
	if err != nil {
		panic(err)
	}

	return cidrNet
}
