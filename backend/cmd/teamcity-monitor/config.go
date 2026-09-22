package main

import (
	"fmt"
	"time"
)

type config struct {
	ServeRESTAddress      string        `env:"SERVE_REST_ADDRESS" envDefault:":8080"`
	ConfigPath            string        `env:"CONFIG_PATH" envDefault:"/app/config.toml"`
	PollInterval          time.Duration `env:"POLL_INTERVAL" envDefault:"20s"`
	InsecureSkipTLSVerify bool          `env:"INSECURE_SKIP_TLS_VERIFY" envDefault:"false"`
}

func (c *config) validate() error {
	if c.PollInterval <= 0 {
		return fmt.Errorf("POLL_INTERVAL must be greater than zero (got %s)", c.PollInterval)
	}
	return nil
}
