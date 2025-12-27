package main

import (
	"github.com/gocloud9/gen-cobra-flags/example"
	"github.com/spf13/cobra"
	"time"
)

type CreateFooBarRequest struct {
	Debug bool
	Duration time.Duration `json:"Duration" yaml:"Duration"`
	Name string `json:"Name" yaml:"Name" validate:"true" example:"custom"`
	Namespace string
	Number int `json:"Number" yaml:"Number"`
	SomeIntMap map[string]int `json:"SomeIntMap" yaml:"SomeIntMap"`
	SomeStingMap map[string]string `json:"SomeStringMap" yaml:"SomeStringMap"`
	Time time.Time `json:"Time" yaml:"Time"`
}

// AddCreateFooBarRequestFlags adds flags for Config to the cobra command
func AddCreateFooBarRequestFlags(cmd *cobra.Command) {
	cmd.Flags().BoolP("debug", "d", false, "Enable debug mode")
	cmd.Flags().DurationP("duration", "T", "5s", "Duration value")
	cmd.Flags().StringP("name", "N", "", "Name of FooBar")
	cmd.Flags().IntP("number", "n", 10, "Number of items")
	cmd.Flags().StringP("some-int-map", "", "{}", "Some int map")
	cmd.Flags().StringToStringP("some-string-map", "", "{}", "Some string map")
	cmd.Flags().StringP("time", "t", time.Now(), "Some Time value")
}

type CreateFooBarRequestOptions struct {
	Debug bool
	Duration time.Duration
	Name string
	Namespace string
	Number int
	SomeIntMap map[string]int
	SomeStingMap map[string]string
	Time time.Time
}

func (c *CreateFooBarRequest) ToCreateFooBarRequest(opts *CreateFooBarRequestOptions) *example.CreateFooBarRequest {
	r := &example.CreateFooBarRequest{
Debug: c.Debug
Duration: c.Duration
Name: c.Name
Namespace: c.Namespace
Number: c.Number
SomeIntMap: c.SomeIntMap
SomeStingMap: c.SomeStingMap
Time: c.Time
}

return r
}

