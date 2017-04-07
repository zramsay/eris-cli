package config

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/monax/cli/version"

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
	IpfsHost          string `json:"IpfsHost,omitempty" yaml:"IpfsHost,omitempty" toml:"IpfsHost,omitempty"`
	IpfsPort          string `json:"IpfsPort,omitempty" yaml:"IpfsPort,omitempty" toml:"IpfsPort,omitempty"`
	CompilersHost     string `json:"CompilersHost,omitempty" yaml:"CompilersHost,omitempty" toml:"CompilersHost,omitempty"` // currently unused
	CompilersPort     string `json:"CompilersPort,omitempty" yaml:"CompilersPort,omitempty" toml:"CompilersPort,omitempty"` // currently unused
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
		errKnown := fmt.Sprintf("List available definitions with the [monax %s ls --known] command", filepath.Base(definitionPath))
		return nil, fmt.Errorf("Unable to find the %q definition: %v\n\n%s", definitionName, os.ErrNotExist, errKnown)
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
	config, err := SetDefaults()
	if err != nil {
		return config, err
	}

	config.AddConfigPath(MonaxRoot)
	config.SetConfigName("monax")
	if err := config.ReadInConfig(); err != nil {
		// Do nothing as this is not essential.
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
	config.SetDefault("CompilersHost", "https://compilers.monax.io")
	config.SetDefault("CompilersPort", "1"+strings.Replace(strings.Split(version.VERSION, "-")[0], ".", "", -1))

	return config, nil
}

// Save writes the "monax.toml" definition file at the default
// location populated by settings.
func Save(settings *Settings) error {
	if settings == nil {
		return fmt.Errorf("cannot save uninitialized settings")
	}

	writer, err := os.Create(filepath.Join(MonaxRoot, "monax.toml"))
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
