package config

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	ver "github.com/eris-ltd/eris-cli/version"

	dir "github.com/eris-ltd/common/go/common"

	"github.com/BurntSushi/toml"
	"github.com/spf13/viper"
	"github.com/tcnksm/go-gitconfig"
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
	CompilersHost  string `json:"CompilersHost,omitempty" yaml:"CompilersHost,omitempty" toml:"CompilersHost,omitempty"` // currently unused
	CompilersPort  string `json:"CompilersPort,omitempty" yaml:"CompilersPort,omitempty" toml:"CompilersPort,omitempty"` // currently unused
	DockerHost     string `json:"DockerHost,omitempty" yaml:"DockerHost,omitempty" toml:"DockerHost,omitempty"`
	DockerCertPath string `json:"DockerCertPath,omitempty" yaml:"DockerCertPath,omitempty" toml:"DockerCertPath,omitempty"`
	CrashReport    string `json:"CrashReport,omitempty" yaml:"CrashReport,omitempty" toml:"CrashReport,omitempty"`
	Verbose        bool

	//image defaults
	ERIS_REG_DEF string `json:"ERIS_REG_DEF,omitempty" yaml:"ERIS_REG_DEF,omitempty" toml:"ERIS_REG_DEF,omitempty"`
	ERIS_REG_BAK string `json:"ERIS_REG_BAK,omitempty" yaml:"ERIS_REG_BAK,omitempty" toml:"ERIS_REG_BAK,omitempty"`

	ERIS_IMG_DATA string `json:"ERIS_IMG_DATA,omitempty" yaml:"ERIS_IMG_DATA,omitempty" toml:"ERIS_IMG_DATA,omitempty"`
	ERIS_IMG_KEYS string `json:"ERIS_IMG_KEYS,omitempty" yaml:"ERIS_IMG_KEYS,omitempty" toml:"ERIS_IMG_KEYS,omitempty"`
	ERIS_IMG_DB   string `json:"ERIS_IMG_DB,omitempty" yaml:"ERIS_IMG_DB,omitempty" toml:"ERIS_IMG_DB,omitempty"`
	ERIS_IMG_PM   string `json:"ERIS_IMG_PM,omitempty" yaml:"ERIS_IMG_PM,omitempty" toml:"ERIS_IMG_PM,omitempty"`
	ERIS_IMG_CM   string `json:"ERIS_IMG_CM,omitempty" yaml:"ERIS_IMG_CM,omitempty" toml:"ERIS_IMG_CM,omitempty"`
	ERIS_IMG_IPFS string `json:"ERIS_IMG_IPFS,omitempty" yaml:"ERIS_IMG_IPFS,omitempty" toml:"ERIS_IMG_IPFS,omitempty"`
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

func LoadViperConfig(configPath, configName string) (*viper.Viper, error) {
	var conf = viper.New()

	conf.AddConfigPath(configPath)
	conf.SetConfigName(configName)
	err := conf.ReadInConfig()
	if err != nil {
		errmsg := fmt.Sprintf(`Unable to load the %q config. Check the file name existence and formatting: %v
`, configName, err)

		switch configPath {
		case dir.ChainsPath, dir.ServicesPath, dir.ActionsPath:
			errknown := fmt.Sprintf(`
List available definitions with the [eris %s ls --known] command
`, filepath.Base(configPath))
			return nil, fmt.Errorf("%s%s", errmsg, errknown)
		}

		return nil, fmt.Errorf(errmsg)
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

	// assorted defaults
	globalConfig.SetDefault("IpfsHost", "http://0.0.0.0") // [csk] TODO: be less opinionated here...
	globalConfig.SetDefault("CrashReport", "bugsnag")

	// compilers defaults
	globalConfig.SetDefault("CompilersHost", "https://compilers.eris.industries")
	verSplit := strings.Split(ver.VERSION, "-")
	verSplit = strings.Split(verSplit[0], ".")
	maj, _ := strconv.Atoi(verSplit[0])
	min, _ := strconv.Atoi(verSplit[1])
	pat, _ := strconv.Atoi(verSplit[2])
	globalConfig.SetDefault("CompilersPort", fmt.Sprintf("1%01d%02d%01d", maj, min, pat))

	// image defaults
	globalConfig.SetDefault("ERIS_REG_DEF", ver.ERIS_REG_DEF)
	globalConfig.SetDefault("ERIS_REG_BAK", ver.ERIS_REG_BAK)
	globalConfig.SetDefault("ERIS_IMG_DATA", ver.ERIS_IMG_DATA)
	globalConfig.SetDefault("ERIS_IMG_KEYS", ver.ERIS_IMG_KEYS)
	globalConfig.SetDefault("ERIS_IMG_DB", ver.ERIS_IMG_DB)
	globalConfig.SetDefault("ERIS_IMG_PM", ver.ERIS_IMG_PM)
	globalConfig.SetDefault("ERIS_IMG_CM", ver.ERIS_IMG_CM)
	globalConfig.SetDefault("ERIS_IMG_COMP", ver.ERIS_IMG_COMP)
	globalConfig.SetDefault("ERIS_IMG_IPFS", ver.ERIS_IMG_IPFS)

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
	if err := enc.Encode(config); err != nil {
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
	case "CrashReport":
		return GlobalConfig.Config.CrashReport
	//image defaults
	case "ERIS_REG_DEF":
		return GlobalConfig.Config.ERIS_REG_DEF
	case "ERIS_REG_BAK":
		return GlobalConfig.Config.ERIS_REG_BAK
	case "ERIS_IMG_DATA":
		return GlobalConfig.Config.ERIS_IMG_DATA
	case "ERIS_IMG_KEYS":
		return GlobalConfig.Config.ERIS_IMG_KEYS
	case "ERIS_IMG_DB":
		return GlobalConfig.Config.ERIS_IMG_DB
	case "ERIS_IMG_PM":
		return GlobalConfig.Config.ERIS_IMG_PM
	case "ERIS_IMG_CM":
		return GlobalConfig.Config.ERIS_IMG_CM
	case "ERIS_IMG_IPFS":
		return GlobalConfig.Config.ERIS_IMG_IPFS
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
	dir.ChainTypePath = filepath.Join(dir.ChainsPath, "chain-types")

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
