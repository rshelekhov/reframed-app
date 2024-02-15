package config

import (
	"github.com/spf13/viper"
	"log"
)

func MustLoad() *ServerSettings {
	cfg := ServerSettings{}

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
