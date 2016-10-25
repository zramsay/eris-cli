package config

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/eris-ltd/eris-cli/version"

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

	// Image defaults.
	DefaultRegistry string `json:"DefaultRegistry,omitempty" yaml:"DefaultRegistry,omitempty" toml:"DefaultRegistry,omitempty"`
	BackupRegistry  string `json:"BackupRegistry,omitempty" yaml:"BackupRegistry,omitempty" toml:"BackupRegistry,omitempty"`

	ImageData      string `json:"ImageData,omitempty" yaml:"ImageData,omitempty" toml:"ImageData,omitempty"`
	ImageKeys      string `json:"ImageKeys,omitempty" yaml:"ImageKeys,omitempty" toml:"ImageKeys,omitempty"`
	ImageDB        string `json:"ImageDB,omitempty" yaml:"ImageDB,omitempty" toml:"ImageDB,omitempty"`
	ImagePM        string `json:"ImagePM,omitempty" yaml:"ImagePM,omitempty" toml:"ImagePM,omitempty"`
	ImageCM        string `json:"ImageCM,omitempty" yaml:"ImageCM,omitempty" toml:"ImageCM,omitempty"`
	ImageIPFS      string `json:"ImageIPFS,omitempty" yaml:"ImageIPFS,omitempty" toml:"ImageIPFS,omitempty"`
	ImageCompilers string `json:"ImageCompilers,omitempty" yaml:"ImageCompilers,omitempty" toml:"ImageCompilers,omitempty"`
}

// New initializes the global configuration with default settings
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

// LoadViper reads the definition file pointed to by
// the definitionPath path and definitionName filename.
func LoadViper(definitionPath, definitionName string) (*viper.Viper, error) {
	var errKnown string
	switch definitionPath {
	case ServicesPath:
		errKnown = fmt.Sprintf(`

List available definitions with the [eris %s ls --known] command`, filepath.Base(definitionPath))
	}

	// Don't use ReadInConfig() for checking file existence because
	// is error is too murky (e.g.:it doesn't say "file not found").
	//
	// Don't use os.Stat() for checking file existence because there might
	// be a selection of supported definition files, e.g.: keys.toml,
	// keys.json, keys.yaml, etc.
	if matches, _ := filepath.Glob(filepath.Join(definitionPath, definitionName+".*")); len(matches) == 0 {
		return nil, fmt.Errorf("Unable to find the %q definition: %v%s", definitionName, os.ErrNotExist, errKnown)
	}

	conf := viper.New()
	conf.AddConfigPath(definitionPath)
	conf.SetConfigName(definitionName)
	if err := conf.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("Unable to load the %q definition: %v%s", definitionName, err, errKnown)
	}

	return conf, nil
}

// Load reads the Viper definition file from the default location.
func Load() (*viper.Viper, error) {
	config, err := SetDefaults()
	if err != nil {
		return config, err
	}

	config.AddConfigPath(ErisRoot)
	config.SetConfigName("eris")
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

	// Image defaults.
	config.SetDefault("DefaultRegistry", version.DefaultRegistry)
	config.SetDefault("BackupRegistry", version.BackupRegistry)
	config.SetDefault("ImageData", version.ImageData)
	config.SetDefault("ImageKeys", version.ImageKeys)
	config.SetDefault("ImageDB", version.ImageDB)
	config.SetDefault("ImagePM", version.ImagePM)
	config.SetDefault("ImageCM", version.ImageCM)
	config.SetDefault("ImageIPFS", version.ImageIPFS)
	config.SetDefault("ImageCompilers", version.ImageCompilers)

	return config, nil
}

// Save writes the "eris.toml" definition file at the default
// location populated by settings.
func Save(settings *Settings) error {
	if settings == nil {
		return fmt.Errorf("cannot save uninitialized settings")
	}

	writer, err := os.Create(filepath.Join(ErisRoot, "eris.toml"))
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
