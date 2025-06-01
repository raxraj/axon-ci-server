package config

import (
	"github.com/spf13/viper"
	"log"
)

func InitConfig() {
	viper.SetConfigName("config") // config file name without extension
	viper.SetConfigType("yaml")   // or "json", "toml", etc.
	viper.AddConfigPath(".")      // look for config in the working directory

	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("fatal error config file: %w", err)
	}
}
