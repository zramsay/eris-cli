package config

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	dir "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/common"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/BurntSushi/toml"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/viper"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/tcnksm/go-gitconfig"
)

// Properly scope the globalConfig.
var GlobalConfig *ErisCli

type ErisCli struct {
	Writer                 io.Writer
	ErrorWriter            io.Writer
	InteractiveWriter      io.Writer
	InteractiveErrorWriter io.Writer
	Config                 *ErisConfig
	ErisDir                string
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
		Writer:                 writer,
		ErrorWriter:            errorWriter,
		InteractiveWriter:      ioutil.Discard,
		InteractiveErrorWriter: ioutil.Discard,
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

func LoadViperConfig(configPath, configName, typ string) (*viper.Viper, error) {
	var conf = viper.New()

	conf.AddConfigPath(configPath)
	conf.SetConfigName(configName)
	err := conf.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("Unable to load the %s's config for %s in %s.\nCheck your known %ss with:\neris %ss ls --known", typ, configName, configPath, typ, typ)
	}

	return conf, nil
}

func LoadGlobalConfig() (*viper.Viper, error) {
	globalConfig, err := SetDefaults()
	if err != nil {
		return globalConfig, err
	}

	globalConfig.AddConfigPath(dir.ErisRoot)
	globalConfig.SetConfigName("eris")
	if err := globalConfig.ReadInConfig(); err != nil {
		// do nothing as this is not essential.
	}

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

// config values will be coerced into strings...
func GetConfigValue(key string) string {
	if GlobalConfig == nil || GlobalConfig.Config == nil {
		return ""
	}

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

// TODO: [by csk] refactor this and DRY it up as this function really should be in common (without the globalConfig object of course)
func ChangeErisDir(erisDir string) {
	if os.Getenv("TESTING") == "true" {
		return
	}

	// Do nothing if not initialized.
	if GlobalConfig == nil {
		return
	}

	GlobalConfig.ErisDir = erisDir
	dir.ErisRoot = erisDir

	// Major directories.
	dir.ActionsPath = filepath.Join(dir.ErisRoot, "actions")
	dir.AppsPath = filepath.Join(dir.ErisRoot, "apps")     // previously "dapps"
	dir.ChainsPath = filepath.Join(dir.ErisRoot, "chains") // previously "blockchains"
	dir.KeysPath = filepath.Join(dir.ErisRoot, "keys")
	dir.RemotesPath = filepath.Join(dir.ErisRoot, "remotes")
	dir.ScratchPath = filepath.Join(dir.ErisRoot, "scratch")
	dir.ServicesPath = filepath.Join(dir.ErisRoot, "services")

	// Chains Directories
	dir.DefaultChainPath = filepath.Join(dir.ChainsPath, "default")
	dir.AccountsTypePath = filepath.Join(dir.ChainsPath, "account-types")

	// Keys Directories
	dir.KeysDataPath = filepath.Join(dir.KeysPath, "data")
	dir.KeyNamesPath = filepath.Join(dir.KeysPath, "names")

	// Scratch Directories (basically eris' cache) (globally coordinated)
	dir.DataContainersPath = filepath.Join(dir.ScratchPath, "data")
	dir.LanguagesScratchPath = filepath.Join(dir.ScratchPath, "languages") // previously "~/.eris/languages"

	// Services Directories
	dir.PersonalServicesPath = filepath.Join(dir.ServicesPath, "global")
}

func marshallGlobalConfig(globalConfig *viper.Viper, config *ErisConfig) error {
	err := globalConfig.Unmarshal(config)
	if err != nil {
		return err
	}

	return nil
}

func GitConfigUser() (uName string, email string, err error) {
	uName, err = gitconfig.Username()
	if err != nil {
		uName = ""
	}
	email, err = gitconfig.Email()
	if err != nil {
		email = ""
	}

	if uName == "" && email == "" {
		err = fmt.Errorf("Can not find username or email in git config. Using \"\" for both\n")
	} else if uName == "" {
		err = fmt.Errorf("Can not find username in git config. Using \"\"\n")
	} else if email == "" {
		err = fmt.Errorf("Can not find email in git config. Using \"\"\n")
	}
	return
}
