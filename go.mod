module github.com/gocloud9/gen-cobra-flags

go 1.26.0

require (
	github.com/gocloud9/gen-cobra-flags/sdk v0.0.0-pre
	github.com/gocloud9/gen-tool v0.0.9
)

replace (
	github.com/gocloud9/gen-cobra-flags/sdk => ./sdk
	github.com/gocloud9/gen-tool => ../gen-tool
)

require (
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/mod v0.30.0 // indirect
	golang.org/x/sync v0.18.0 // indirect
	golang.org/x/tools v0.39.0 // indirect
)
