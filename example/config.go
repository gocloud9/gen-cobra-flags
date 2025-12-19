package example

//go:generate gen-cobra-flags -input config.go -struct Config -output flags_gen.go -package example

// Config represents application configuration
type Config struct {
	Host     string `flag:"host" short:"H" usage:"Server host" default:"localhost"`
	Port     int    `flag:"port" short:"p" usage:"Server port" default:"8080"`
	Debug    bool   `flag:"debug" short:"d" usage:"Enable debug mode"`
	Verbose  bool   `flag:"verbose" short:"v" usage:"Enable verbose logging"`
	LogLevel string `flag:"log-level" short:"l" usage:"Log level (info, warn, error)" default:"info"`
}
