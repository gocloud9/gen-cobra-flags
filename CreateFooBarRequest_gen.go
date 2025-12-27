// Generated Code - Do Not Edit
package todo

import (
    "time"
    "net"
)

type CreateFooBarRequest struct {
    CIDR net.IPNet `json:"IPAddress" yaml:"IPAddress"`
    Debug bool
    Duration time.Duration `json:"Duration" yaml:"Duration"`
    IPAddress net.IP `json:"IPAddress" yaml:"IPAddress"`
    MyCustomType MyCustomType `json:"MyCustomType" yaml:"MyCustomType"`
    Name string `json:"Name" yaml:"Name" validate:"true" example:"custom"`
    Namespace string
    Number int `json:"Number" yaml:"Number"`
    SomeIntMap map[string]int `json:"SomeIntMap" yaml:"SomeIntMap"`
    SomeStingMap map[string]string `json:"SomeStringMap" yaml:"SomeStringMap"`
    Time time.Time `json:"Time" yaml:"Time"`
}

// AddCreateFooBarRequestFlags adds flags for Config to the cobra command
func AddCreateFooBarRequestFlags(cmd *cobra.Command) {
	cmd.Flags().IPNetP("ip-address", "", "", "IP Address")
	cmd.Flags().BoolP("debug", "d", false, "Enable debug mode")
	cmd.Flags().DurationP("duration", "T", "5s", "Duration value")
	cmd.Flags().IPP("ip-address", "", "", "IP Address")
	cmd.Flags().stringP("my-custom-type", "", "", "Some custom type")
	cmd.Flags().StringP("name", "N", "", "Name of FooBar")
	cmd.Flags().IntP("number", "n", 10, "Number of items")
	cmd.Flags().StringToIntP("some-int-map", "", "{}", "Some int map")
	cmd.Flags().StringToStringP("some-string-map", "", "{}", "Some string map")
	cmd.Flags().TimeP("time", "t", time.Now(), "Some Time value")
}

type CreateFooBarRequestOptions struct {
    Adaptors map[string]func(any) any
    Namespace *string
    Number *int
}

func (c *CreateFooBarRequest) ToCreateFooBarRequest(opts ...*CreateFooBarRequestOptions) *example.CreateFooBarRequest {
    r := &example.CreateFooBarRequest {
        CIDR: c.CIDR,
        Debug: c.Debug,
        Duration: c.Duration,
        IPAddress: c.IPAddress,
        MyCustomType: c.MyCustomType,
        Name: c.Name,
        Namespace: c.Namespace,
        Number: c.Number,
        SomeIntMap: c.SomeIntMap,
        SomeStingMap: c.SomeStingMap,
        Time: c.Time,
    }

    for i := range opts {
      if opts[i].Namespace != nil {
          r.Namespace = *opts[i].Namespace
      }
      if opts[i].Number != nil {
          r.Number = *opts[i].Number
      }

    }

    return r
}
