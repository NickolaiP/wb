package config

import "github.com/caarlos0/env/v10"

type Config struct {
	Host     string `env:"POSTGRES_HOST" envDefault:"localhost"`
	User     string `env:"POSTGRES_USER" envDefault:"user"`
	Password string `env:"POSTGRES_PASSWORD" envDefault:"1234"`
	Port     int    `env:"POSTGRES_PORT" envDefault:"5432"`
	Db       string `env:"POSTGRES_DB" envDefault:"test"`
}

func InitConfig() (Config, error) {
	var config Config
	// парсим переменные среды
	err := env.Parse(&config)
	if err != nil {
		return Config{}, err
	}

	return config, nil
}
