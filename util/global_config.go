package util

import (
	"io"
	"os"
	"path/filepath"

	dir "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/BurntSushi/toml"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/viper"
)

// Properly scope the globalConfig
var GlobalConfig *ErisCli

type ErisCli struct {
	Writer      io.Writer
	ErrorWriter io.Writer
	Config      *ErisConfig
}

type ErisConfig struct {
	IpfsHost       string `json:"IpfsHost,omitempty" yaml:"IpfsHost,omitempty" toml:"IpfsHost,omitempty"`
	CompilersHost  string `json:"CompilersHost,omitempty" yaml:"CompilersHost,omitempty" toml:"CompilersHost,omitempty"`
	DockerHost     string `json:"DockerHost,omitempty" yaml:"DockerHost,omitempty" toml:"DockerHost,omitempty"`
	DockerCertPath string `json:"DockerCertPath,omitempty" yaml:"DockerCertPath,omitempty" toml:"DockerCertPath,omitempty"`

	Verbose bool
}

func SetGlobalObject(writer, errorWriter io.Writer) (*ErisCli, error) {
	e := ErisCli{
		Writer:      writer,
		ErrorWriter: errorWriter,
	}

	config, err := LoadGlobalConfig()
	if err != nil {
		return &e, err
	}

	e.Config = &ErisConfig{}

	err = marshallGlobalConfig(config, e.Config)
	if err != nil {
		return &e, err
	}
	return &e, nil
}

func LoadGlobalConfig() (*viper.Viper, error) {
	globalConfig, err := SetDefaults()
	if err != nil {
		return globalConfig, err
	}

	globalConfig.AddConfigPath(dir.ErisRoot)
	globalConfig.SetConfigName("eris")
	globalConfig.ReadInConfig()

	return globalConfig, nil
}

func SetDefaults() (*viper.Viper, error) {
	var globalConfig = viper.New()
	globalConfig.SetDefault("IpfsHost", "http://0.0.0.0")
	globalConfig.SetDefault("CompilersHost", "https://compilers.eris.industries")
	return globalConfig, nil
}

func SaveGlobalConfig(config *ErisConfig) error {
	writer, err := os.Create(filepath.Join(dir.ErisRoot, "eris.toml"))
	defer writer.Close()
	if err != nil {
		return err
	}

	enc := toml.NewEncoder(writer)
	enc.Indent = ""
	err = enc.Encode(config)
	if err != nil {
		return err
	}
	return nil
}

func marshallGlobalConfig(globalConfig *viper.Viper, config *ErisConfig) error {
	err := globalConfig.Marshal(config)
	if err != nil {
		return err
	}

	return nil
}
