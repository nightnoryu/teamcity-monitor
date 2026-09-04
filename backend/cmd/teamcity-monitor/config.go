package main

type config struct {
	ServeRESTAddress string `env:"SERVE_REST_ADDRESS" envDefault:":8080"`
	ConfigPath       string `env:"CONFIG_PATH" envDefault:"/app/config.toml"`
}
