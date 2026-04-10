package blaze

// ServerOption configures a Server.
type ServerOption func(*serverConfig)

type serverConfig struct {
	port      int
	logOutput LogOutput
	logFile   string
}

func defaultConfig() serverConfig {
	return serverConfig{
		port:      0, // random free port
		logOutput: LogStdout,
	}
}

// WithPort sets the port the server listens on. Default is 0 (random free port).
func WithPort(port int) ServerOption {
	return func(c *serverConfig) {
		c.port = port
	}
}

// WithLogOutput sets where logs are written. Default is LogStdout.
// When using LogFile or LogBoth, you must also call WithLogFile.
func WithLogOutput(output LogOutput) ServerOption {
	return func(c *serverConfig) {
		c.logOutput = output
	}
}

// WithLogFile sets the file path for log output. Used with LogFile or LogBoth.
func WithLogFile(path string) ServerOption {
	return func(c *serverConfig) {
		c.logFile = path
	}
}
