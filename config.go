package main

import (
	"fmt"

	"github.com/caarlos0/env"
)

// Config is a placeholder for available configurations
type Config struct {
	Port       int    `env:"PORT" envDefault:"8080"`
	MysqlDSN   string `env:"MYSQL_DSN" envDefault:"guest:guest@tcp(guest:3306)/bmu"`
	LegacySalt string `env:"LEGACY_SALT,required"`
	JWTSecret  string `env:"JWT_SECRET,required"`
	WebHost    string `env:"WEB_HOST" envDefault:"http://localhost:8080"`
}

var config Config

func init() {
	config = Config{}
	err := env.Parse(&config)
	if err != nil {
		fmt.Printf("%+v\n", err)
	}
}
