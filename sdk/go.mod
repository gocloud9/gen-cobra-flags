module sdk

go 1.24.11

require (
	github.com/gocloud9/gen-cobra-flags/sdk main
)

replace (
	github.com/gocloud9/gen-cobra-flags/sdk => ./sdk
)
