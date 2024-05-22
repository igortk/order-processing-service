package config

import (
	"github.com/caarlos0/env/v6"
	"order-processing-service/util"
)

type (
	Config struct {
		RabbitConfig RabbitConfig `envPrefix:"RMQ_"`
		LoggerConfig LoggerConfig `envPrefix:"LOG_"`
		RedisConfig  RedisConfig  `envPrefix:"RS_"`
	}

	RabbitConfig struct {
		Host              string `env:"HOST" envDefault:"localhost"`
		Port              int    `env:"PORT"  envDefault:"5672"`
		Username          string `env:"USERNAME" envDefault:"guest"`
		Password          string `env:"PASSWORD" envDefault:"guest"`
		VirtualHost       string `env:"VIRTUAL_HOST" envDefault:"/"`
		ReconnectAttempts uint   `env:"RECONNECT_ATTEMPTS" envDefault:"5"`
	}

	LoggerConfig struct {
		Level string `env:"LEVEL" envDefault:"INFO"`
	}

	RedisConfig struct {
		Host     string `env:"HOST" envDefault:"localhost"`
		Port     int    `env:"PORT"  envDefault:"6379"`
		Password string `env:"PASSWORD" envDefault:""`
		Database uint   `env:"DB" envDefault:"1"`
	}
)

func GetConfig() *Config {
	var cfg Config
	err := env.Parse(&cfg)
	util.IsError(err, "Can't parse config")

	return &cfg
}
