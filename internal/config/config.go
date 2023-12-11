package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"os"
	"time"
)

type Config struct {
	Env        string           `yaml:"env" env-default:"development"`
	HTTPServer HTTPServerConfig `yaml:"http_server"`
	Postgres   PostgresConfig   `yaml:"postgres" env-required:"true"`
}

type HTTPServerConfig struct {
	Address     string        `yaml:"address" env-default:"0.0.0.0:8080"`
	Timeout     time.Duration `yaml:"timeout" env-default:"5s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
}

type PostgresConfig struct {
	DBName     string `yaml:"db_name"`
	DBTestName string `yaml:"db_test_name"`
	Host       string `yaml:"host"`
	Port       string `yaml:"port"`
	User       string `yaml:"user"`
	Password   string `yaml:"password"`
	SSLMode    string `yaml:"sslmode" env-default:"disable"`
}

func MustLoad() *Config {
	// Get the path to the config file from the env variable CONFIG_PATH
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("CONFIG_PATH environment variable is not set")
	}

	// Check if the config file exists
	if _, err := os.Stat(configPath); err != nil {
		log.Fatalf("error opening config file: %s", err)
	}

	var cfg Config

	// Read the config file and fill in the struct
	err := cleanenv.ReadConfig(configPath, &cfg)
	if err != nil {
		log.Fatalf("error reading config file: %s", err)
	}

	return &cfg
}
