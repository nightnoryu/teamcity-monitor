package main

import (
	"fmt"
	"os"
	"time"

	"github.com/nightnoryu/go-kita/env"
	"github.com/nightnoryu/go-kita/jsonlog"
)

type config struct {
	ServeRESTAddress      string        `env:"SERVE_REST_ADDRESS" envDefault:":8080"`
	MetricsAddress        string        `env:"METRICS_ADDRESS"`
	LogLevel              jsonlog.Level `env:"LOG_LEVEL" envDefault:"info"`
	ConfigPath            string        `env:"CONFIG_PATH" envDefault:"/app/config.toml"`
	PollInterval          time.Duration `env:"POLL_INTERVAL" envDefault:"20s"`
	InsecureSkipTLSVerify bool          `env:"INSECURE_SKIP_TLS_VERIFY" envDefault:"false"`
}

func loadConfig() (*config, error) {
	cfg, err := env.ParseEnv[config](appID)
	if err != nil {
		return nil, err
	}
	// An explicitly empty value disables the listener.
	if _, set := os.LookupEnv("TEAMCITY_MONITOR_METRICS_ADDRESS"); !set {
		cfg.MetricsAddress = "127.0.0.1:9090"
	}
	return cfg, nil
}

func (c *config) validate() error {
	if c.PollInterval <= 0 {
		return fmt.Errorf("POLL_INTERVAL must be greater than zero (got %s)", c.PollInterval)
	}
	return nil
}
