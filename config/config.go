package config

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	ver "github.com/eris-ltd/eris-cli/version"

	dir "github.com/eris-ltd/common/go/common"

	"github.com/BurntSushi/toml"
	"github.com/spf13/viper"
	gitconfig "github.com/tcnksm/go-gitconfig"
)

// Global carries CLI settings global across packages.
var Global *Config

// Config describes CLI global settings.
type Config struct {
	Writer                 io.Writer
	ErrorWriter            io.Writer
	InteractiveWriter      io.Writer
	InteractiveErrorWriter io.Writer
	Settings
}

// Settings describes settings loadable from "eris.toml"
// configuration file.
type Settings struct {
	IpfsHost       string `json:"IpfsHost,omitempty" yaml:"IpfsHost,omitempty" toml:"IpfsHost,omitempty"`
	IpfsPort       string `json:"IpfsPort,omitempty" yaml:"IpfsPort,omitempty" toml:"IpfsPort,omitempty"`
	CompilersHost  string `json:"CompilersHost,omitempty" yaml:"CompilersHost,omitempty" toml:"CompilersHost,omitempty"` // currently unused
	CompilersPort  string `json:"CompilersPort,omitempty" yaml:"CompilersPort,omitempty" toml:"CompilersPort,omitempty"` // currently unused
	DockerHost     string `json:"DockerHost,omitempty" yaml:"DockerHost,omitempty" toml:"DockerHost,omitempty"`
	DockerCertPath string `json:"DockerCertPath,omitempty" yaml:"DockerCertPath,omitempty" toml:"DockerCertPath,omitempty"`
	CrashReport    string `json:"CrashReport,omitempty" yaml:"CrashReport,omitempty" toml:"CrashReport,omitempty"`
	ImagesPullTimeout string `json:"ImagesPullTimeout,omitempty" yaml:"ImagesPullTimeout,omitempty" toml:"ImagesPullTimeout,omitempty"`
	Verbose        bool

	// Image defaults.
	DefaultRegistry string `json:"DefaultRegistry,omitempty" yaml:"DefaultRegistry,omitempty" toml:"DefaultRegistry,omitempty"`
	BackupRegistry  string `json:"BackupRegistry,omitempty" yaml:"BackupRegistry,omitempty" toml:"BackupRegistry,omitempty"`

	ImageData string `json:"ImageData,omitempty" yaml:"ImageData,omitempty" toml:"ImageData,omitempty"`
	ImageKeys string `json:"ImageKeys,omitempty" yaml:"ImageKeys,omitempty" toml:"ImageKeys,omitempty"`
	ImageDB   string `json:"ImageDB,omitempty" yaml:"ImageDB,omitempty" toml:"ImageDB,omitempty"`
	ImagePM   string `json:"ImagePM,omitempty" yaml:"ImagePM,omitempty" toml:"ImagePM,omitempty"`
	ImageCM   string `json:"ImageCM,omitempty" yaml:"ImageCM,omitempty" toml:"ImageCM,omitempty"`
	ImageIPFS string `json:"ImageIPFS,omitempty" yaml:"ImageIPFS,omitempty" toml:"ImageIPFS,omitempty"`
}

// New initializes the global config with default settings
// or settings loaded from the "eris.toml" default location.
// New also initialize default writer and errorWriter streams.
// Viper or unmarshalling errors are returned on error.
func New(writer, errorWriter io.Writer) (*Config, error) {
	config := &Config{
		Writer:                 writer,
		ErrorWriter:            errorWriter,
		InteractiveWriter:      ioutil.Discard,
		InteractiveErrorWriter: ioutil.Discard,
	}

	v, err := Load()
	if err != nil {
		return config, err
	}

	if err := v.Unmarshal(&config.Settings); err != nil {
		return config, err
	}

	return config, nil
}

// LoadViper reads the configuration file pointed to by
// the configPath path and configName filename. 
func LoadViper(configPath, configName string) (*viper.Viper, error) {
	var errKnown string
	switch configPath {
	case dir.ChainsPath, dir.ServicesPath, dir.ActionsPath:
		errKnown = fmt.Sprintf(`

List available definitions with the [eris %s ls --known] command`, filepath.Base(configPath))
	}

	// Don't use ReadInConfig() for checking file existence because
	// is error is too murky (e.g.:it doesn't say "file not found").
	//
	// Don't use os.Stat() for checking file existence because there might
	// be a selection of supported definition files, e.g.: keys.toml,
	// keys.json, keys.yaml, etc.
	if matches, _ := filepath.Glob(filepath.Join(configPath, configName+".*")); len(matches) == 0 {
		return nil, fmt.Errorf("Unable to find the %q definition: %v%s", configName, os.ErrNotExist, errKnown)
	}

	conf := viper.New()
	conf.AddConfigPath(configPath)
	conf.SetConfigName(configName)
	if err := conf.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("Unable to load the %q definition: %v%s", configName, err, errKnown)
	}

	return conf, nil
}

// Load reads the Viper configuration file from the default location.
func Load() (*viper.Viper, error) {
	config, err := SetDefaults()
	if err != nil {
		return config, err
	}

	config.AddConfigPath(dir.ErisRoot)
	config.SetConfigName("eris")
	if err := config.ReadInConfig(); err != nil {
		// do nothing as this is not essential.
	}

	return config, nil
}

// SetDefaults initializes the Viper struct with default settings.
func SetDefaults() (*viper.Viper, error) {
	var config = viper.New()

	config.SetDefault("IpfsHost", "http://0.0.0.0") // [csk] TODO: be less opinionated here...
	config.SetDefault("IpfsPort", "8080")           // [csk] TODO: be less opinionated here...
	config.SetDefault("CrashReport", "bugsnag")
	config.SetDefault("ImagesPullTimeout", "15m")

	// Compiler defaults.
	config.SetDefault("CompilersHost", "https://compilers.eris.industries")
	config.SetDefault("CompilersPort", "1"+strings.Replace(strings.Split(ver.VERSION, "-")[0], ".", "", -1))

	// Image defaults.
	config.SetDefault("DefaultRegistry", ver.DefaultRegistry)
	config.SetDefault("BackupRegistry", ver.BackupRegistry)
	config.SetDefault("ImageData", ver.ImageData)
	config.SetDefault("ImageKeys", ver.ImageKeys)
	config.SetDefault("ImageDB", ver.ImageDB)
	config.SetDefault("ImagePM", ver.ImagePM)
	config.SetDefault("ImageCM", ver.ImageCM)
	config.SetDefault("ImageIPFS", ver.ImageIPFS)

	return config, nil
}

// Save writes the "eris.toml" configuration file at the default
// location populated by settings.
func Save(settings *Settings) error {
	writer, err := os.Create(filepath.Join(dir.ErisRoot, "eris.toml"))
	defer writer.Close()
	if err != nil {
		return err
	}

	enc := toml.NewEncoder(writer)
	enc.Indent = ""
	if err := enc.Encode(settings); err != nil {
		return err
	}
	return nil
}

// GitConfigUser returns Git global settings of the current user.
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
