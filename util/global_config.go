package util

import (
	// "os"
	// "fmt"
	"io"

	dir "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/viper"
)

type ErisCli struct {
	Writer         *io.Writer
	ErrorWriter    *io.Writer
	Config         *ErisConfig
}

type ErisConfig struct {
	IpfsHost       string        `json:"," yaml:"," toml:","`
	DockerHost     string        `json:"," yaml:"," toml:","`
	DockerCertPath string        `json:"," yaml:"," toml:","`
	CompilersHost  string        `json:"," yaml:"," toml:","`

	Verbose        bool
}

func SetGlobalObject(globalConfig *viper.Viper, writer, errorWriter *io.Writer) (*ErisCli, error) {
	e := &ErisCli{
		Writer:      writer,
		ErrorWriter: errorWriter,
		// Config:      LoadGlobalConfig(),
	}

	return e, nil
}

func LoadGlobalConfig() error {
	// SetDefaultSettings(globalConfig)
	globalConfig := viper.New()
	globalConfig.AddConfigPath(dir.ErisRoot)
	globalConfig.SetConfigName("config")
	// err := globalConfig.ReadInConfig()
	// if err != nil {
	// 	return err
	// }
	return nil
}

func SetDefaultSettings(globalConfig *viper.Viper) {
	// globalConfig.SetDefault("1234", false)

}
