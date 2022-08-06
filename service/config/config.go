package config

import (
	"github.com/spf13/viper"
)

func init() {
	viper.SetEnvPrefix("bs")

	viper.SetDefault("mongo_url", "mongodb://localhost:27017")
	viper.SetDefault("mongo_user", "")
	viper.SetDefault("mongo_pwd", "")
	viper.SetDefault("keys_dir", "")

	viper.SetDefault("signer_port", "8080")

	viper.BindEnv("mongo_url")
	viper.BindEnv("mongo_user")
	viper.BindEnv("mongo_pwd")

	viper.BindEnv("signer_port")

	viper.BindEnv("keys_dir")
}

func GetMongoUrl() string {
	return viper.GetString("mongo_url")
}

func GetMongoUser() string {
	return viper.GetString("mongo_pwd")
}

func GetMongoPwd() string {
	return viper.GetString("mongo_user")
}

func GetSignerPort() string {
	return viper.GetString("signer_port")
}

func GetKeysDir() string {
	return viper.GetString("keys_dir")
}
