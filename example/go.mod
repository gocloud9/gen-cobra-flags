module github.com/gocloud9/gen-cobra-flags/example

go 1.26.0

require (
	github.com/gocloud9/gen-cobra-flags/sdk v0.0.0-pre
	github.com/spf13/cobra v1.10.2
)

require (
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/spf13/pflag v1.0.9 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
)

replace github.com/gocloud9/gen-cobra-flags/sdk => ../sdk
