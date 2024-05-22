package config

import (
	"log"

	"github.com/spf13/viper"
)

func Init() {

	viperInit()
}

var config *viper.Viper
var schemas *viper.Viper

func viperInit() {
	config = viper.New()
	schemas = viper.New()
	readEnv(config, "env", "json", "./config")
}

func readEnv(config *viper.Viper, fileName string, fileType string, location string) {
	*config = *(viper.New())
	(*config).SetConfigName(fileName)
	(*config).SetConfigType(fileType)
	(*config).AddConfigPath(location)

	if err := (*config).ReadInConfig(); err != nil {
		log.Fatal("fatal error config file")
		panic(err)
	}
}

func GetConfig() *viper.Viper {
	return config
}

func GetConfigProp(key string) string {
	return config.GetString(key)
}

func GetSchemas() *viper.Viper {
	return schemas
}

func GetSchemasProp(key string) string {
	return schemas.GetString(key)
}
