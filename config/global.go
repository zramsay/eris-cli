package config

import (
	"fmt"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/viper"
)

func PlopEntireConfig(globalConfig *viper.Viper, args []string) {
	for _, key := range args {
		fmt.Println(globalConfig.GetString(key))
	}
}

func Set(args []string) {

}

func Edit() {

}
