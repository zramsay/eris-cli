package config

import (
	"fmt"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/viper"
)

func PlopEntireConfig(globalConfig *viper.Viper, args []string) {
	for _, arg := range args {
		fmt.Printf("%s -> %s\n", arg, globalConfig.GetString(arg))
	}
}

func Set(args []string) {

}

func Edit() {

}
