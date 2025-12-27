package example

import (
	"net"
	"time"
)

//go:generate gen-cobra-flags -input config.go -struct Config -output flags_gen.go -package example

type MyCustomType int

// +cobra:flag=config
// +cobra:short=c
// +cobra:usage=Configuration for the server
// +cobra:default={}
// +cobra:adaptorName=CreateFooBarConfig
type CreateFooBarRequest struct {
	// +cobra:flag=name
	// +cobra:short=N
	// +cobra:usage=Name of FooBar
	// +cobra:json=Name
	// +cobra:yaml=Name
	// +cobra:required=true
	// +cobra:default=""
	// +cobra:customTags=validate:"true" example:"custom"
	Name string

	// +cobra:option=true
	Namespace string

	// +cobra:flag=number
	// +cobra:short=n
	// +cobra:usage=Number of items
	// +cobra:default=10
	// +cobra:json=Number
	// +cobra:yaml=Number
	// +cobra:option=true
	Number int

	// +cobra:flag=time
	// +cobra:short=t
	// +cobra:usage=Some Time value
	// +cobra:default=time.Now()
	// +cobra:json=Time
	// +cobra:yaml=Time
	Time time.Time

	// +cobra:flag=duration
	// +cobra:short=T
	// +cobra:usage=Duration value
	// +cobra:default="5s"
	// +cobra:json=Duration
	// +cobra:yaml=Duration
	Duration time.Duration

	// +cobra:flag=some-string-map
	// +cobra:usage=Some string map
	// +cobra:default="{}"
	// +cobra:json=SomeStringMap
	// +cobra:yaml=SomeStringMap
	SomeStingMap map[string]string

	// +cobra:flag=some-int-map
	// +cobra:usage=Some int map
	// +cobra:default="{}"
	// +cobra:json=SomeIntMap
	// +cobra:yaml=SomeIntMap
	SomeIntMap map[string]int

	// +cobra:flag=debug
	// +cobra:short=d
	// +cobra:usage=Enable debug mode
	// +cobra:default=false
	Debug bool

	// +cobra:flag=ip-address
	// +cobra:usage=IP Address
	// +cobra:default=""
	// +cobra:json=IPAddress
	// +cobra:yaml=IPAddress
	IPAddress net.IP

	// +cobra:flag=ip-address
	// +cobra:usage=IP Address
	// +cobra:default=""
	// +cobra:json=IPAddress
	// +cobra:yaml=IPAddress
	CIDR net.IPNet

	// +cobra:flag=my-custom-type
	// +cobra:usage=Some custom type
	// +cobra:default=""
	// +cobra:json=MyCustomType
	// +cobra:yaml=MyCustomType
	// +cobra:adaptor=MyCustomTypeAdaptor
	MyCustomType MyCustomType
}
