package config

import "time"

type ServerSettings struct {
	AppEnv     string             `mapstructure:"APP_ENV"`
	AppData    AppDataSettings    `mapstructure:",squash"`
	HTTPServer HTTPServerSettings `mapstructure:",squash"`
	Postgres   PostgresSettings   `mapstructure:",squash"`
	Clients    ClientsSettings    `mapstructure:",squash"`
}

type AppDataSettings struct {
	ID      string `mapstructure:"APP_ID"`
	Name    string `mapstructure:"APP_NAME"`
	Secret  string `mapstructure:"APP_SECRET"`
	BaseURL string `mapstructure:"APP_BASE_URL"`
}

type HTTPServerSettings struct {
	Address          string        `mapstructure:"HTTP_SERVER_ADDRESS"`
	Timeout          time.Duration `mapstructure:"HTTP_SERVER_TIMEOUT" envDefault:"10s"`
	IdleTimeout      time.Duration `mapstructure:"HTTP_SERVER_IDLE_TIMEOUT" envDefault:"60s"`
	RequestLimitByIP int           `mapstructure:"HTTP_SERVER_REQUEST_LIMIT_BY_IP" envDefault:"100"`
}

type PostgresSettings struct {
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

type ClientsSettings struct {
	SSO Client `mapstructure:",squash"`
}

type Client struct {
	Address      string        `mapstructure:"SSO_CLIENT_ADDRESS"`
	Timeout      time.Duration `mapstructure:"SSO_CLIENT_TIMEOUT"`
	RetriesCount int           `mapstructure:"SSO_CLIENT_RETRIES_COUNT"`
	// TODO: implement secure transport
	// Insecure     bool          `mapstructure:"SSO_CLIENT_INSECURE"`
}
