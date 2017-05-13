package main

import (
	"fmt"

	"github.com/caarlos0/env"
)

// Config is a placeholder for available configurations
type Config struct {
	Port int `env:"PORT" envDefault:"8080"`
}

// EnvConfig returns a config struct with values prepopulated from ENV
func EnvConfig() Config {
	cfg := Config{}
	err := env.Parse(&cfg)
	if err != nil {
		fmt.Printf("%+v\n", err)
	}
	return cfg
}
