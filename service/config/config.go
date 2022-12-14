package config

import (
	"github.com/spf13/viper"
	"os"
)

func init() {
	viper.SetEnvPrefix("bs")

	viper.SetDefault("mongo_url", "mongodb://localhost:27017")
	viper.SetDefault("mongo_user", "")
	viper.SetDefault("mongo_pwd", "")
	viper.SetDefault("keys_dir", "")

	viper.SetDefault("enable_mongo_xact", false)

	viper.SetDefault("msg_signer_url", "http://localhost:8080")
	viper.SetDefault("signer_port", "8080")

	// total signers env variable
	// when stateful set is used this is set to total
	// number of signing pods
	// for local development this is 1
	viper.SetDefault("total_signers", 1)

	viper.SetDefault("batch_size", 100)

	// signer id is identifier for the current pod
	// we adapt k8s format e.g. <signer name>-0, <signer name>-2, ...
	viper.SetDefault("my_pod_name", "signer-0")

	// fault injection for test purposes in percents
	viper.SetDefault("test_sign_failure_rate_pct", 0)

	// generate-record tool
	viper.SetDefault("record_generator_batch_size", 5000)
	// in 16 byte multiple, e.g. record_generator_message_size_16=16 is 256 bytes
	viper.SetDefault("record_generator_message_size_16", 16)

	viper.BindEnv("mongo_url")
	viper.BindEnv("mongo_user")
	viper.BindEnv("mongo_pwd")

	viper.BindEnv("msg_signer_url")
	viper.BindEnv("signer_port")

	viper.BindEnv("keys_dir")

	viper.BindEnv("enable_mongo_xact")

	viper.BindEnv("total_signers")
	viper.BindEnv("batch_size")
	viper.BindEnv("my_pod_name")
	viper.BindEnv("test_sign_failure_rate_pct")

	// generate-record tool
	viper.BindEnv("record_generator_batch_size")
	viper.BindEnv("record_generator_message_size_16")
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

func GetMsgSignerUrl() string {
	return viper.GetString("msg_signer_url")
}

func GetSignerPort() string {
	return viper.GetString("signer_port")
}

func GetKeysDir() string {
	return viper.GetString("keys_dir")
}

func GetEnableMongoXact() bool {
	return viper.GetBool("enable_mongo_xact")
}

func GetTotalSigners() int {
	return viper.GetInt("total_signers")
}

func GetTestSignFailureRatePct() int {
	return viper.GetInt("test_sign_failure_rate_pct")
}

func GetMyPodName() string {
	// in k8s pod name is available as HOSTNAME
	hostname := os.Getenv("HOSTNAME")
	if hostname == "" {
		return viper.GetString("my_pod_name")
	}
	return hostname
}

func GetBatchSize() int {
	return viper.GetInt("batch_size")
}

// generate-record tool
func GetRecordGeneratorBatchSize() int {
	return viper.GetInt("record_generator_batch_size")
}

func GetRecordGeneratorMessageSize16() int {
	return viper.GetInt("record_generator_message_size_16")
}
