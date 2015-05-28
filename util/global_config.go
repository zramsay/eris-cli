package util

import (
	"fmt"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/viper"
)

func SetDefaultSettings(globalConfig *viper.Viper) {
	globalConfig.SetDefault("1234", false)
}

func LoadGlobalConfig(globalConfig *viper.Viper) {
	SetDefaultSettings(globalConfig)
	globalConfig.AddConfigPath(UserErisDir())
	globalConfig.SetConfigType("json")
	globalConfig.SetConfigName("config")
	err := globalConfig.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
}
