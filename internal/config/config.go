package config

import (
	"github.com/spf13/viper"
	"log"
	"time"
)

type Config struct {
	AppEnv     string           `mapstructure:"APP_ENV"`
	HTTPServer HTTPServerConfig `mapstructure:",squash"`
	Postgres   PostgresConfig   `mapstructure:",squash"`
}

type HTTPServerConfig struct {
	Address     string        `mapstructure:"HTTP_SERVER_ADDRESS"`
	Timeout     time.Duration `mapstructure:"HTTP_SERVER_TIMEOUT"`
	IdleTimeout time.Duration `mapstructure:"HTTP_SERVER_IDLE_TIMEOUT"`
}

type PostgresConfig struct {
	DBName   string `mapstructure:"DB_NAME"`
	Host     string `mapstructure:"DB_HOST"`
	Port     string `mapstructure:"DB_PORT"`
	User     string `mapstructure:"DB_USER"`
	Password string `mapstructure:"DB_PASSWORD"`
	SSLMode  string `mapstructure:"DB_SSL_MODE"`
	URL      string `mapstructure:"DB_URL"`
}

func MustLoad() *Config {
	cfg := Config{}

	viper.SetConfigFile(".env")

	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("error finding or reading config file: %s", err)
	}

	viper.AutomaticEnv()

	err = viper.Unmarshal(&cfg)
	if err != nil {
		log.Fatalf("error unmarshalling config file into struct: %s: ", err)
	}

	return &cfg
}
