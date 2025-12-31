module github.com/gocloud9/gen-cobra-flags

go 1.24.11

require (
	github.com/gocloud9/gen-cobra-flags/example v0.0.0-20251227234251-a843dc22fc99
	github.com/gocloud9/gen-cobra-flags/sdk v0.0.0-pre
	github.com/gocloud9/gen-tool v0.0.7
	github.com/spf13/cobra v1.10.2
)

replace github.com/gocloud9/gen-cobra-flags/sdk => ./sdk

require (
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/spf13/pflag v1.0.9 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/mod v0.30.0 // indirect
	golang.org/x/sync v0.18.0 // indirect
	golang.org/x/tools v0.39.0 // indirect
)
