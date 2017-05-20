package config

import (
	"fmt"

	"github.com/caarlos0/env"
)

// Configuration is a placeholder for available configurations
type Configuration struct {
	Port       int    `env:"PORT" envDefault:"8080"`
	MysqlDSN   string `env:"MYSQL_DSN" envDefault:"boxmeup:boxmeup@tcp(localhost:3306)/boxmeup"`
	LegacySalt string `env:"LEGACY_SALT,required"`
	JWTSecret  string `env:"JWT_SECRET,required"`
	WebHost    string `env:"WEB_HOST" envDefault:"http://localhost:8080"`
}

var Config Configuration

func init() {
	Config = Configuration{}
	err := env.Parse(&Config)
	if err != nil {
		fmt.Printf("%+v\n", err)
	}
}
