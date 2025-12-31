// Generated Code with gen-cobra-flags - Do Not Edit
package generated

import (
	"fmt"
	"github.com/gocloud9/gen-cobra-flags/example"
	"github.com/gocloud9/gen-cobra-flags/sdk/pkg/adaptors"
	"github.com/spf13/cobra"
	"net"
	"time"
)

type CreateFooBarRequestConfig struct {
	AConversionOfTypes int32
	CIDR               *net.IPNet `json:"IPAddress" yaml:"IPAddress"`
	Debug              bool
	Duration           time.Duration `json:"Duration" yaml:"Duration"`
	IPAddress          net.IP        `json:"IPAddress" yaml:"IPAddress"`
	MyCustomType       MyCustomType  `json:"MyCustomType" yaml:"MyCustomType"`
	Name               string        `json:"Name" yaml:"Name" validate:"true" example:"custom"`
	Namespace          string
	Number             int               `json:"Number" yaml:"Number"`
	SomeIntMap         map[string]int    `json:"SomeIntMap" yaml:"SomeIntMap"`
	SomeStingMap       map[string]string `json:"SomeStringMap" yaml:"SomeStringMap"`
	Time               time.Time         `json:"Time" yaml:"Time"`
}

// AddCreateFooBarRequestFlags adds flags for Config to the cobra command
func AddCreateFooBarRequestFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("config", "c", "", "Configuration for the server")
	cmd.Flags().Int64P("a-conversion-of-types", "", 1, "A conversion of types")
	cmd.Flags().IPNetP("ip-address", "", net.IPNet{}, "IP Address")
	cmd.Flags().BoolP("debug", "d", false, "Enable debug mode")
	cmd.Flags().DurationP("duration", "T", time.Duration(0), "Duration value")
	cmd.Flags().IPP("ip-address", "", net.IP{}, "IP Address")
	cmd.Flags().StringP("my-custom-type", "", "", "Some custom type")
	cmd.Flags().StringP("name", "N", "", "Name of FooBar")
	cmd.Flags().IntP("number", "n", 10, "Number of items")
	cmd.Flags().StringToIntP("some-int-map", "", nil, "Some int map")
	cmd.Flags().StringToStringP("some-string-map", "", nil, "Some string map")
	cmd.Flags().TimeP("time", "t", time.Now(), []string{time.RFC3339}, "Some Time value")
}

type CreateFooBarRequestOptions struct {
	Namespace *string
	Number    *int
}

func CreateFooBarRequestConfigFromFlags(cmd *cobra.Command) (*CreateFooBarRequestConfig, error) {
	cin, err := cmd.Flags().GetString("config")
	if err != nil {
		return nil, fmt.Errorf("getting CreateFooBarRequest config from flags: %w", err)
	}

	c, err := adaptors.JsonOrYamlToStruct[CreateFooBarRequestConfig]([]byte(cin))
	if err != nil {
		return nil, fmt.Errorf("adapting CreateFooBarRequest config from flags: %w", err)
	}
	c.AConversionOfTypes, err = cmd.Flags().GetString("a-conversion-of-types")
	if err != nil {
		return nil, fmt.Errorf("getting flag AConversionOfTypes: %w", err)
	}

	c.CIDR, err = cmd.Flags().GetIPNet("ip-address")
	if err != nil {
		return nil, fmt.Errorf("getting flag CIDR: %w", err)
	}

	c.Debug, err = cmd.Flags().GetBool("debug")
	if err != nil {
		return nil, fmt.Errorf("getting flag Debug: %w", err)
	}

	c.Duration, err = cmd.Flags().GetDuration("duration")
	if err != nil {
		return nil, fmt.Errorf("getting flag Duration: %w", err)
	}

	c.IPAddress, err = cmd.Flags().GetIP("ip-address")
	if err != nil {
		return nil, fmt.Errorf("getting flag IPAddress: %w", err)
	}

	c.MyCustomType, err = cmd.Flags().GetString("my-custom-type")
	if err != nil {
		return nil, fmt.Errorf("getting flag MyCustomType: %w", err)
	}

	c.Name, err = cmd.Flags().GetString("name")
	if err != nil {
		return nil, fmt.Errorf("getting flag Name: %w", err)
	}

	c.Number, err = cmd.Flags().GetInt("number")
	if err != nil {
		return nil, fmt.Errorf("getting flag Number: %w", err)
	}

	c.SomeIntMap, err = cmd.Flags().GetStringToInt("some-int-map")
	if err != nil {
		return nil, fmt.Errorf("getting flag SomeIntMap: %w", err)
	}

	c.SomeStingMap, err = cmd.Flags().GetStringToString("some-string-map")
	if err != nil {
		return nil, fmt.Errorf("getting flag SomeStingMap: %w", err)
	}

	c.Time, err = cmd.Flags().GetTime("time")
	if err != nil {
		return nil, fmt.Errorf("getting flag Time: %w", err)
	}

	return &c, nil
}

func (c *CreateFooBarRequestConfig) ToCreateFooBarRequest(opts ...*CreateFooBarRequestOptions) (*example.CreateFooBarRequest, error) {
	r := &example.CreateFooBarRequest{}

	var err error

	r.AConversionOfTypes, err = adaptorCustomInt32ToString(c.AConversionOfTypes)
	if err != nil {
		return nil, fmt.Errorf("adapting field AConversionOfTypes: %w", err)
	}

	r.CIDR = c.CIDR

	r.Debug = c.Debug

	r.Duration = c.Duration

	r.IPAddress = c.IPAddress

	r.MyCustomType = c.MyCustomType

	r.Name = c.Name

	r.Namespace = c.Namespace

	r.Number = c.Number

	r.SomeIntMap = c.SomeIntMap

	r.SomeStingMap = c.SomeStingMap

	r.Time = c.Time

	for i := range opts {
		if opts[i].Namespace != nil {
			r.Namespace = *opts[i].Namespace
		}
		if opts[i].Number != nil {
			r.Number = *opts[i].Number
		}

	}

	return r, nil
}
