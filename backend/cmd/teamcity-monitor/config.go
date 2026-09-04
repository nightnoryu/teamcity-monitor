package main

type config struct {
	ServeRESTAddress string `env:"SERVE_REST_ADDRESS" envDefault:":8080"`
}
