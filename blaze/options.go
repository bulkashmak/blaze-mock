package blaze

// ServerOption configures a Server.
type ServerOption func(*serverConfig)

type serverConfig struct {
	port int
}

func defaultConfig() serverConfig {
	return serverConfig{
		port: 0, // random free port
	}
}

// WithPort sets the port the server listens on. Default is 0 (random free port).
func WithPort(port int) ServerOption {
	return func(c *serverConfig) {
		c.port = port
	}
}
