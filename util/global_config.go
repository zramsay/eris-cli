package util

import (
	"io"
	"os"
	"path"
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
	globalConfig.SetDefault("IpfsHost", "http://0.0.0.0:8080")
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

// config values will be coerced into strings...
func GetConfigValue(key string) string {
	switch key {
	case "IpfsHost":
		return GlobalConfig.Config.IpfsHost
	case "CompilersHost":
		return GlobalConfig.Config.CompilersHost
	case "DockerHost":
		return GlobalConfig.Config.DockerHost
	case "DockerCertPath":
		return GlobalConfig.Config.DockerCertPath
	default:
		return ""
	}
}

func ChangeErisDir(erisDir string) {
	if os.Getenv("TEST_IN_CIRCLE") == "true" {
		return
	}

	dir.ErisRoot = erisDir

	// Major Directories
	dir.ActionsPath = path.Join(dir.ErisRoot, "actions")
	dir.BlockchainsPath = path.Join(dir.ErisRoot, "blockchains")
	dir.DataContainersPath = path.Join(dir.ErisRoot, "data")
	dir.DappsPath = path.Join(dir.ErisRoot, "dapps")
	dir.FilesPath = path.Join(dir.ErisRoot, "files")
	dir.KeysPath = path.Join(dir.ErisRoot, "keys")
	dir.LanguagesPath = path.Join(dir.ErisRoot, "languages")
	dir.ServicesPath = path.Join(dir.ErisRoot, "services")
	dir.ScratchPath = path.Join(dir.ErisRoot, "scratch")

	// Scratch Directories (globally coordinated)
	dir.EpmScratchPath = path.Join(dir.ScratchPath, "epm")
	dir.LllcScratchPath = path.Join(dir.ScratchPath, "lllc")
	dir.SolcScratchPath = path.Join(dir.ScratchPath, "sol")
	dir.SerpScratchPath = path.Join(dir.ScratchPath, "ser")
}

func marshallGlobalConfig(globalConfig *viper.Viper, config *ErisConfig) error {
	err := globalConfig.Marshal(config)
	if err != nil {
		return err
	}

	return nil
}
