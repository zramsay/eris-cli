package config

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

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

// Settings describes settings loadable from "monax.toml"
// definition file.
type Settings struct {
	DockerHost        string `json:"DockerHost,omitempty" yaml:"DockerHost,omitempty" toml:"DockerHost,omitempty"`
	DockerCertPath    string `json:"DockerCertPath,omitempty" yaml:"DockerCertPath,omitempty" toml:"DockerCertPath,omitempty"`
	CrashReport       string `json:"CrashReport,omitempty" yaml:"CrashReport,omitempty" toml:"CrashReport,omitempty"`
	ImagesPullTimeout string `json:"ImagesPullTimeout,omitempty" yaml:"ImagesPullTimeout,omitempty" toml:"ImagesPullTimeout,omitempty"`
	Verbose           bool
}

// New initializes the global configuration with default settings
// or settings loaded from the "monax.toml" default location.
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

// LoadViper reads the definition file pointed to by
// the definitionPath path and definitionName filename.
func LoadViper(definitionPath, definitionName string) (*viper.Viper, error) {

	// Don't use ReadInConfig() for checking file existence because
	// is error is too murky (e.g.:it doesn't say "file not found").
	//
	// Don't use os.Stat() for checking file existence because there might
	// be a selection of supported definition files, e.g.: keys.toml,
	// keys.json, keys.yaml, etc.

	if matches, _ := filepath.Glob(filepath.Join(definitionPath, definitionName+".*")); len(matches) == 0 {
		return nil, fmt.Errorf("Unable to find the %q definition: %v", definitionName, os.ErrNotExist)
	}

	conf := viper.New()
	conf.AddConfigPath(definitionPath)
	conf.SetConfigName(definitionName)
	if err := conf.ReadInConfig(); err != nil {
		// [zr] this error to deduplicate with loaders/services.go:66 in #468
		return nil, fmt.Errorf("Formatting error with your %q definition:\n\n%v", definitionName, err)
	}

	return conf, nil
}

// Load reads the Viper definition file from the default location.
func Load() (*viper.Viper, error) {
	var config = viper.New()

	config.SetDefault("ImagesPullTimeout", "15m")
	config.AddConfigPath(MonaxRoot)
	config.SetConfigName("monax")
	_ = config.ReadInConfig()

	return config, nil
}

// Save writes the "monax.toml" definition file at the default
// location populated by settings.
func Save(settings *Settings) error {
	if settings == nil {
		return fmt.Errorf("cannot save uninitialized settings")
	}

	writer, err := os.Create(filepath.Join(MonaxRoot, "monax.toml"))
	if err != nil {
		return err
	}
	defer writer.Close()

	enc := toml.NewEncoder(writer)
	enc.Indent = ""
	return enc.Encode(settings)
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
