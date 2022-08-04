package config

import (
	"github.com/spf13/viper"
)

func init() {
	viper.SetEnvPrefix("bs")
	viper.SetDefault("mongo_url", "mongodb://localhost:27017")
	viper.SetDefault("signer_port", "8080")

	viper.BindEnv("mongo_url")
	viper.BindEnv("signer_port")
}

func GetMongoUrl() string {
	return viper.GetString("mongo_url")
}

func GetSignerPort() string {
	return viper.GetString("signer_port")
}
