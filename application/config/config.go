package config

import (
	"bytes"
	"fmt"

	"github.com/spf13/viper"
)

func Init() {
	viperInit()
}

var config *viper.Viper
var userConfig *viper.Viper

func viperInit() {
	config = viper.New()
	userConfig = viper.New()
	readEnv(config, "env", "json", "./config")
}

func readEnv(config *viper.Viper, fileName string, fileType string, location string) {
	*config = *(viper.New())
	(*config).SetConfigName(fileName)
	(*config).SetConfigType(fileType)
	(*config).AddConfigPath(location)

	if err := (*config).ReadInConfig(); err != nil {
		panic(fmt.Errorf("could not read config: %s", err))
	}
}

func GetConfig() *viper.Viper {
	return config
}

func GetConfigProp(key string) string {
	return config.GetString(key)
}

func GetUserConfig() *viper.Viper {
	return userConfig
}

func GetUserConfigProp(key string) string {
	return userConfig.GetString(key)
}

func SetUserConfig(json string) error {
	return userConfig.ReadConfig(bytes.NewBufferString(json))
}
