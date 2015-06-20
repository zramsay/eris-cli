package util

import (
	// "os"
	// "fmt"

  dir "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/viper"
)

func SetDefaultSettings(globalConfig *viper.Viper) {
	globalConfig.SetDefault("1234", false)
}

func LoadGlobalConfig(globalConfig *viper.Viper) {
	SetDefaultSettings(globalConfig)
	globalConfig.AddConfigPath(dir.ErisRoot)
	globalConfig.SetConfigType("json")
	globalConfig.SetConfigName("config")
	// err := globalConfig.ReadInConfig()
	// if err != nil {
	// 	fmt.Println("Fatal error config file ->\n  %v", err)
	// 	os.Exit(1)
	// }
}
