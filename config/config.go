package config

import (
	"github.com/spf13/viper"
	"log"
	"time"
)

type (
	Config struct {
		AppEnv     string           `mapstructure:"APP_ENV"`
		HTTPServer HTTPServerConfig `mapstructure:",squash"`
		Postgres   PostgresConfig   `mapstructure:",squash"`
		JWTAuth    JWTConfig        `mapstructure:",squash"`
	}

	HTTPServerConfig struct {
		Address     string        `mapstructure:"HTTP_SERVER_ADDRESS"`
		Timeout     time.Duration `mapstructure:"HTTP_SERVER_TIMEOUT" envDefault:"10s"`
		IdleTimeout time.Duration `mapstructure:"HTTP_SERVER_IDLE_TIMEOUT" envDefault:"60s"`
	}

	PostgresConfig struct {
		Host string `mapstructure:"DB_HOST" envDefault:"localhost"`
		Port string `mapstructure:"DB_PORT" envDefault:"5432"`

		DBName   string `mapstructure:"DB_NAME"`
		User     string `mapstructure:"DB_USER"`
		Password string `mapstructure:"DB_PASSWORD"`

		SSLMode string `mapstructure:"DB_SSL_MODE" envDefault:"disable"`
		ConnURL string `mapstructure:"DB_CONN_URL"`

		ConnPoolSize int           `mapstructure:"DB_CONN_POOL_SIZE" envDefault:"10"`
		ReadTimeout  time.Duration `mapstructure:"DB_READ_TIMEOUT" envDefault:"5s"`
		WriteTimeout time.Duration `mapstructure:"DB_WRITE_TIMEOUT" envDefault:"5s"`
		IdleTimeout  time.Duration `mapstructure:"DB_IDLE_TIMEOUT" envDefault:"60s"`
		DialTimeout  time.Duration `mapstructure:"DB_DIAL_TIMEOUT" envDefault:"10s"`
	}

	JWTConfig struct {
		Secret                 string        `mapstructure:"JWT_SECRET"`
		AccessTokenTTL         time.Duration `mapstructure:"JWT_ACCESS_TOKEN_TTL"`
		RefreshTokenTTL        time.Duration `mapstructure:"JWT_REFRESH_TOKEN_TTL"`
		RefreshTokenCookiePath string        `mapstructure:"JWT_REFRESH_TOKEN_COOKIE_PATH"`
	}
)

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
