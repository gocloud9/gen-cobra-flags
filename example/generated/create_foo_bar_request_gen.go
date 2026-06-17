// Generated Code with gen-cobra-flags - Do Not Edit
package generated

import (
    "time"
    "net"
    "github.com/spf13/cobra"
)

type CreateFooBarRequestConfig struct {
    AConversionOfTypes int32
    CIDR *net.IPNet `json:"CIDR" yaml:"CIDR"`
    Debug bool
    Duration time.Duration `json:"Duration" yaml:"Duration"`
    IPAddress net.IP `json:"IPAddress" yaml:"IPAddress"`
    MyCustomType example.MyCustomType `json:"MyCustomType" yaml:"MyCustomType"`
    Name string `json:"Name" yaml:"Name" validate:"true" example:"custom"`
    Namespace string
    Number int `json:"Number" yaml:"Number"`
    SomeIntMap map[string]int `json:"SomeIntMap" yaml:"SomeIntMap"`
    SomeStingMap map[string]string `json:"SomeStringMap" yaml:"SomeStringMap"`
    Time time.Time `json:"Time" yaml:"Time"`
}

// AddCreateFooBarRequestFlags adds flags for Config to the cobra command
func AddCreateFooBarRequestFlags(cmd *cobra.Command) {
    cmd.Flags().StringP("config", "c", "", "Configuration for the server")
    cmd.Flags().Int64P("a-conversion-of-types", "", 1, "A conversion of types")
    cmd.Flags().IPNetP("cidr", "", net.IPNet{}, "CIDR BLOCK")
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
    Number *int
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

    if c.AConversionOfTypes == 0 {
        c.AConversionOfTypes, err = adaptorCustomInt64ToInt32(1)
        if err != nil {
            return nil, fmt.Errorf("adapting default value for flag a-conversion-of-types: %w", err)
        }
    }

    inAConversionOfTypes, err := cmd.Flags().GetInt64("a-conversion-of-types")
    if err != nil {
        return nil, fmt.Errorf("getting flag a-conversion-of-types: %w", err)
    }
    if inAConversionOfTypes != 1 {
        c.AConversionOfTypes, err = adaptorCustomInt64ToInt32(inAConversionOfTypes)
        if err != nil {
            return nil, fmt.Errorf("adapting flag a-conversion-of-types: %w", err)
        }
    }
    if reflect.DeepEqual(c.CIDR, nil) {
        c.CIDR, err = adaptors.ToPtr(net.IPNet{})
        if err != nil {
            return nil, fmt.Errorf("adapting default value for flag cidr: %w", err)
        }
    }

    inCIDR, err := cmd.Flags().GetIPNet("cidr")
    if err != nil {
        return nil, fmt.Errorf("getting flag cidr: %w", err)
    }
    if reflect.DeepEqual(inCIDR, net.IPNet{}) {
        c.CIDR, err = adaptors.ToPtr(inCIDR)
        if err != nil {
            return nil, fmt.Errorf("adapting flag cidr: %w", err)
        }
    }

    if c.Debug == false {
        c.Debug = false
    }

    inDebug, err := cmd.Flags().GetBool("debug")
    if err != nil {
        return nil, fmt.Errorf("getting flag debug: %w", err)
    }

    if inDebug != false {
       c.Debug = inDebug
    }
    if reflect.DeepEqual(c.Duration, nil) {
        c.Duration = time.Duration(0)
    }

    inDuration, err := cmd.Flags().GetDuration("duration")
    if err != nil {
        return nil, fmt.Errorf("getting flag duration: %w", err)
    }

    if inDuration != time.Duration(0) {
       c.Duration = inDuration
    }
    if reflect.DeepEqual(c.IPAddress, nil) {
        c.IPAddress = net.IP{}
    }

    inIPAddress, err := cmd.Flags().GetIP("ip-address")
    if err != nil {
        return nil, fmt.Errorf("getting flag ip-address: %w", err)
    }

    if inIPAddress != net.IP{} {
       c.IPAddress = inIPAddress
    }

    if c.MyCustomType == 0 {
        c.MyCustomType, err = adaptorStringToMyCustomType("")
        if err != nil {
            return nil, fmt.Errorf("adapting default value for flag my-custom-type: %w", err)
        }
    }

    inMyCustomType, err := cmd.Flags().GetString("my-custom-type")
    if err != nil {
        return nil, fmt.Errorf("getting flag my-custom-type: %w", err)
    }
    if inMyCustomType != "" {
        c.MyCustomType, err = adaptorStringToMyCustomType(inMyCustomType)
        if err != nil {
            return nil, fmt.Errorf("adapting flag my-custom-type: %w", err)
        }
    }

    if c.Name == "" {
        c.Name = ""
    }

    inName, err := cmd.Flags().GetString("name")
    if err != nil {
        return nil, fmt.Errorf("getting flag name: %w", err)
    }

    if inName != "" {
       c.Name = inName
    }

    if c.Number == 0 {
        c.Number = 10
    }

    inNumber, err := cmd.Flags().GetInt("number")
    if err != nil {
        return nil, fmt.Errorf("getting flag number: %w", err)
    }

    if inNumber != 10 {
       c.Number = inNumber
    }
    if reflect.DeepEqual(c.SomeIntMap, nil) {
        c.SomeIntMap = nil
    }

    inSomeIntMap, err := cmd.Flags().GetStringToInt("some-int-map")
    if err != nil {
        return nil, fmt.Errorf("getting flag some-int-map: %w", err)
    }

    if inSomeIntMap != nil {
       c.SomeIntMap = inSomeIntMap
    }
    if reflect.DeepEqual(c.SomeStingMap, nil) {
        c.SomeStingMap = nil
    }

    inSomeStingMap, err := cmd.Flags().GetStringToString("some-string-map")
    if err != nil {
        return nil, fmt.Errorf("getting flag some-string-map: %w", err)
    }

    if inSomeStingMap != nil {
       c.SomeStingMap = inSomeStingMap
    }
    if reflect.DeepEqual(c.Time, nil) {
        c.Time = time.Now()
    }

    inTime, err := cmd.Flags().GetTime("time")
    if err != nil {
        return nil, fmt.Errorf("getting flag time: %w", err)
    }

    if inTime != time.Now() {
       c.Time = inTime
    }
    return &c, nil
}

func (c *CreateFooBarRequestConfig) ToCreateFooBarRequest(opts ...*CreateFooBarRequestOptions) (*example.CreateFooBarRequest, error){
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
      if opts[i].Namespace != nil {r.Namespace = *opts[i].Namespace
      }
      if opts[i].Number != nil {r.Number = *opts[i].Number
      }
    }

    return r, nil
}
