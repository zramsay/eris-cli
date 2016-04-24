package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	log "github.com/Sirupsen/logrus"
	logger "github.com/eris-ltd/common/go/log"
)

var (
	configErisDir = filepath.Join(os.TempDir(), "config")
)

func TestMain(m *testing.M) {
	log.SetFormatter(logger.ConsoleFormatter(log.DebugLevel))

	log.SetLevel(log.ErrorLevel)
	// log.SetLevel(log.InfoLevel)
	// log.SetLevel(log.DebugLevel)

	// Unset this variable by default for config package.
	savedEnv := os.Getenv("TESTING")
	if err := os.Unsetenv("TESTING"); err != nil {
		panic("can't unset TESTING")
	}
	defer os.Setenv("TESTING", savedEnv)

	log.WithField("dir", configErisDir).Info("Using temporary directory for config files")

	m.Run()
}

func TestSetGlobalObject1(t *testing.T) {
	cli, err := SetGlobalObject(os.Stdout, os.Stderr)
	if err != nil {
		t.Fatalf("expected success, got error %v", err)
	}

	if cli.Writer != os.Stdout {
		t.Fatalf("Writer doesn't match os.Stdout")
	}

	if cli.ErrorWriter != os.Stderr {
		t.Fatalf("ErrorWriter doesn't match os.Stderr")
	}
}

func TestSetGlobalObject2(t *testing.T) {
	cli, err := SetGlobalObject(os.Stderr, os.Stdout)
	if err != nil {
		t.Fatalf("expected success, got error %v", err)
	}

	if cli.Writer != os.Stderr {
		t.Fatalf("Writer doesn't match os.Stderr")
	}

	if cli.ErrorWriter != os.Stdout {
		t.Fatalf("ErrorWriter doesn't match os.Stdout")
	}
}

func TestSetGlobalObjectNil(t *testing.T) {
	cli, err := SetGlobalObject(nil, nil)
	if err != nil {
		t.Fatalf("expected success, got error %v", err)
	}

	if cli.Writer != nil {
		t.Fatalf("Writer doesn't match nil")
	}

	if cli.ErrorWriter != nil {
		t.Fatalf("ErrorWriter doesn't match nil")
	}
}

func TestSetGlobalObjectDefaultConfig(t *testing.T) {
	ChangeErisDir(configErisDir)

	cli, err := SetGlobalObject(os.Stderr, os.Stdout)
	if err != nil {
		t.Fatalf("expected success, got error %v", err)
	}

	defaults, err := SetDefaults()
	if err != nil {
		t.Fatalf("expected defaults loaded, got error %v", err)
	}

	if def, returned := defaults.Get("IpfsHost"), cli.Config.IpfsHost; reflect.DeepEqual(returned, def) != true {
		t.Fatalf("expected default %q, got %q", returned, def)
	}

	if def, returned := defaults.Get("CompilersHost"), cli.Config.CompilersHost; reflect.DeepEqual(returned, def) != true {
		t.Fatalf("expected default %q, got %q", returned, def)
	}

	log.WithFields(log.Fields{
		"ipfshost":       cli.Config.IpfsHost,
		"compilers host": cli.Config.CompilersHost,
		"host":           cli.Config.DockerHost,
		"cert path":      cli.Config.DockerCertPath,
		"crash report":   cli.Config.CrashReport,
		"verbose":        cli.Config.Verbose,
	}).Info("Checking defaults")
}

func TestSetGlobalObjectCustomConfig(t *testing.T) {
	placeErisConfig(`
IpfsHost = "foo"
CompilersHost = "bar"
DockerHost = "baz"
DockerCertPath = "qux"
CrashReport = "quux"
Verbose = true
`)
	defer removeErisDir()

	// [pv]: this is a bit awkward way to reinitialize the config:
	// SetGlobalConfig, then ChangeErisDir, then again SetGlobalConfig.
	GlobalConfig = &ErisCli{}
	ChangeErisDir(configErisDir)
	cli, err := SetGlobalObject(os.Stderr, os.Stdout)
	if err != nil {
		t.Fatalf("expected success, got error %v", err)
	}

	if custom, returned := "foo", cli.Config.IpfsHost; custom != returned {
		t.Fatalf("expected %q, got %q", custom, returned)
	}
	if custom, returned := "bar", cli.Config.CompilersHost; custom != returned {
		t.Fatalf("expected %q, got %q", custom, returned)
	}
	if custom, returned := "baz", cli.Config.DockerHost; custom != returned {
		t.Fatalf("expected %q, got %q", custom, returned)
	}
	if custom, returned := "qux", cli.Config.DockerCertPath; custom != returned {
		t.Fatalf("expected %q, got %q", custom, returned)
	}
	if custom, returned := "quux", cli.Config.CrashReport; custom != returned {
		t.Fatalf("expected %q, got %q", custom, returned)
	}
	if custom, returned := true, cli.Config.Verbose; custom != returned {
		t.Fatalf("expected %v, got %v", custom, returned)
	}
}

func TestSetGlobalObjectCustomEmptyConfig(t *testing.T) {
	placeErisConfig(``)
	defer removeErisDir()

	GlobalConfig = &ErisCli{}
	ChangeErisDir(configErisDir)
	cli, err := SetGlobalObject(os.Stderr, os.Stdout)
	if err != nil {
		t.Fatalf("expected success, got error %v", err)
	}

	defaults, err := SetDefaults()
	if err != nil {
		t.Fatalf("expected defaults loaded, got error %v", err)
	}

	log.WithFields(log.Fields{
		"ipfshost":       cli.Config.IpfsHost,
		"compilers host": cli.Config.CompilersHost,
		"host":           cli.Config.DockerHost,
		"cert path":      cli.Config.DockerCertPath,
		"crash report":   cli.Config.CrashReport,
		"verbose":        cli.Config.Verbose,
	}).Info("Checking empty values")

	// With an empty config, the values are used are defaults.
	if def, returned := defaults.Get("IpfsHost"), cli.Config.IpfsHost; reflect.DeepEqual(returned, def) != true {
		t.Fatalf("expected default %v, got %v", returned, def)
	}

	if def, returned := defaults.Get("CompilersHost"), cli.Config.CompilersHost; reflect.DeepEqual(returned, def) != true {
		t.Fatalf("expected default %q, got %q", returned, def)
	}

	if custom, returned := "", cli.Config.DockerHost; custom != returned {
		t.Fatalf("expected %q, got %q", custom, returned)
	}
	if custom, returned := "", cli.Config.DockerCertPath; custom != returned {
		t.Fatalf("expected %q, got %q", custom, returned)
	}
	if custom, returned := "bugsnag", cli.Config.CrashReport; custom != returned {
		t.Fatalf("expected %q, got %q", custom, returned)
	}
	if custom, returned := false, cli.Config.Verbose; custom != returned {
		t.Fatalf("expected %v, got %v", custom, returned)
	}
}

func TestSetGlobalObjectCustomBadConfig(t *testing.T) {
	placeErisConfig(`*`)
	defer removeErisDir()

	GlobalConfig = &ErisCli{}
	ChangeErisDir(configErisDir)
	cli, err := SetGlobalObject(os.Stderr, os.Stdout)
	if err != nil {
		t.Fatalf("expected success, got error %v", err)
	}

	log.WithFields(log.Fields{
		"ipfshost":       cli.Config.IpfsHost,
		"compilers host": cli.Config.CompilersHost,
		"host":           cli.Config.DockerHost,
		"cert path":      cli.Config.DockerCertPath,
		"crash report":   cli.Config.CrashReport,
		"verbose":        cli.Config.Verbose,
	}).Info("Checking empty values")

	// With an empty config, the values are used are defaults.
	defaults, err := SetDefaults()
	if err != nil {
		t.Fatalf("expected defaults loaded, got error %v", err)
	}

	if def, returned := defaults.Get("IpfsHost"), cli.Config.IpfsHost; reflect.DeepEqual(returned, def) != true {
		t.Fatalf("expected default %q, got %q", returned, def)
	}

	if def, returned := defaults.Get("CompilersHost"), cli.Config.CompilersHost; reflect.DeepEqual(returned, def) != true {
		t.Fatalf("expected default %q, got %q", returned, def)
	}

	if custom, returned := "", cli.Config.DockerHost; custom != returned {
		t.Fatalf("expected %q, got %q", custom, returned)
	}
	if custom, returned := "", cli.Config.DockerCertPath; custom != returned {
		t.Fatalf("expected %q, got %q", custom, returned)
	}
	if custom, returned := "bugsnag", cli.Config.CrashReport; custom != returned {
		t.Fatalf("expected %q, got %q", custom, returned)
	}
	if custom, returned := false, cli.Config.Verbose; custom != returned {
		t.Fatalf("expected %v, got %v", custom, returned)
	}
}

func TestSetDefaults(t *testing.T) {
	defaults, err := SetDefaults()
	if err != nil {
		t.Fatalf("expected success, got error %v", err)
	}

	if _, ok := defaults.Get("IpfsHost").(string); !ok {
		t.Fatalf("expected IpfsHost value set")
	}

	if _, ok := defaults.Get("CompilersHost").(string); !ok {
		t.Fatalf("expected CompilersHost values set")
	}
}

func TestLoadGlobalConfig(t *testing.T) {
	placeErisConfig(`
IpfsHost = "foo"
CompilersHost = "bar"
DockerHost = "baz"
DockerCertPath = "qux"
CrashReport = "quux"
Verbose = true
`)
	defer removeErisDir()

	config, err := LoadGlobalConfig()
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	if expected, returned := "foo", config.Get("IpfsHost"); reflect.DeepEqual(expected, returned) != true {
		t.Fatalf("expected %q, got %q", expected, returned)
	}
	if expected, returned := "bar", config.Get("CompilersHost"); reflect.DeepEqual(expected, returned) != true {
		t.Fatalf("expected %q, got %q", expected, returned)
	}
	if expected, returned := "baz", config.Get("DockerHost"); reflect.DeepEqual(expected, returned) != true {
		t.Fatalf("expected %q, got %q", expected, returned)
	}
	if expected, returned := "qux", config.Get("DockerCertPath"); reflect.DeepEqual(expected, returned) != true {
		t.Fatalf("expected %q, got %q", expected, returned)
	}
	if expected, returned := "quux", config.Get("CrashReport"); reflect.DeepEqual(expected, returned) != true {
		t.Fatalf("expected %q, got %q", expected, returned)
	}
	if expected, returned := true, config.Get("Verbose"); reflect.DeepEqual(expected, returned) != true {
		t.Fatalf("expected %v, got %v", expected, returned)
	}
}

func TestLoadGlobalConfigEmpty(t *testing.T) {
	placeErisConfig(``)
	defer removeErisDir()

	config, err := LoadGlobalConfig()
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	defaults, err := SetDefaults()
	if err != nil {
		t.Fatalf("expected defaults loaded, got error %v", err)
	}

	if def, returned := defaults.Get("IpfsHost"), config.Get("IpfsHost"); reflect.DeepEqual(returned, def) != true {
		t.Fatalf("expected default %q, got %q", returned, def)
	}
	if def, returned := defaults.Get("CompilersHost"), config.Get("CompilersHost"); reflect.DeepEqual(returned, def) != true {
		t.Fatalf("expected default %q, got %q", returned, def)
	}
	if returned := config.Get("DockerHost"); returned != nil {
		t.Fatalf("expected nil, got %q", returned)
	}
	if returned := config.Get("DockerCertPath"); returned != nil {
		t.Fatalf("expected nil, got %q", returned)
	}
	if def, returned := config.Get("CrashReport"), config.Get("CrashReport"); reflect.DeepEqual(returned, def) != true {
		t.Fatalf("expected default %q, got %q", returned, def)
	}
	if returned := config.Get("Verbose"); returned != nil {
		t.Fatalf("expected nil, got %q", returned)
	}
}

func TestLoadGlobalConfigBad(t *testing.T) {
	placeErisConfig(`*`)
	defer removeErisDir()

	// With bad config, load defaults.
	config, err := LoadGlobalConfig()
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	defaults, err := SetDefaults()
	if err != nil {
		t.Fatalf("expected defaults loaded, got error %v", err)
	}

	if def, returned := defaults.Get("IpfsHost"), config.Get("IpfsHost"); reflect.DeepEqual(returned, def) != true {
		t.Fatalf("expected default %q, got %q", returned, def)
	}
	if def, returned := defaults.Get("CompilersHost"), config.Get("CompilersHost"); reflect.DeepEqual(returned, def) != true {
		t.Fatalf("expected default %q, got %q", returned, def)
	}
	if returned := config.Get("DockerHost"); returned != nil {
		t.Fatalf("expected nil, got %q", returned)
	}
	if returned := config.Get("DockerCertPath"); returned != nil {
		t.Fatalf("expected nil, got %q", returned)
	}
	if def, returned := config.Get("CrashReport"), config.Get("CrashReport"); reflect.DeepEqual(returned, def) != true {
		t.Fatalf("expected default %q, got %q", returned, def)
	}
	if returned := config.Get("Verbose"); returned != nil {
		t.Fatalf("expected nil, got %q", returned)
	}
}

func TestLoadViperConfig(t *testing.T) {
	placeErisConfig(`
IpfsHost = "foo"
CompilersHost = "bar"
DockerHost = "baz"
DockerCertPath = "qux"
CrashReport = "quux"
Verbose = true
`)
	defer removeErisDir()

	config, err := LoadViperConfig(configErisDir, "eris", "test")
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	if expected, returned := "foo", config.Get("IpfsHost"); reflect.DeepEqual(expected, returned) != true {
		t.Fatalf("expected %q, got %q", expected, returned)
	}
	if expected, returned := "bar", config.Get("CompilersHost"); reflect.DeepEqual(expected, returned) != true {
		t.Fatalf("expected %q, got %q", expected, returned)
	}
	if expected, returned := "baz", config.Get("DockerHost"); reflect.DeepEqual(expected, returned) != true {
		t.Fatalf("expected %q, got %q", expected, returned)
	}
	if expected, returned := "qux", config.Get("DockerCertPath"); reflect.DeepEqual(expected, returned) != true {
		t.Fatalf("expected %q, got %q", expected, returned)
	}
	if expected, returned := "quux", config.Get("CrashReport"); reflect.DeepEqual(expected, returned) != true {
		t.Fatalf("expected %q, got %q", expected, returned)
	}
	if expected, returned := true, config.Get("Verbose"); reflect.DeepEqual(expected, returned) != true {
		t.Fatalf("expected %v, got %v", expected, returned)
	}
}

func TestLoadViperConfigEmpty(t *testing.T) {
	placeErisConfig(``)
	defer removeErisDir()

	config, err := LoadViperConfig(configErisDir, "eris", "test")
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	if returned := config.Get("IpfsHost"); returned != nil {
		t.Fatalf("expected nil, got %q", returned)
	}
	if returned := config.Get("CompilersHost"); returned != nil {
		t.Fatalf("expected nil, got %q", returned)
	}
	if returned := config.Get("DockerHost"); returned != nil {
		t.Fatalf("expected nil, got %q", returned)
	}
	if returned := config.Get("DockerCertPath"); returned != nil {
		t.Fatalf("expected nil, got %q", returned)
	}
	if returned := config.Get("CrashReport"); returned != nil {
		t.Fatalf("expected nil, got %q", returned)
	}
	if returned := config.Get("Verbose"); returned != nil {
		t.Fatalf("expected nil, got %q", returned)
	}
}

func TestLoadViperConfigBad(t *testing.T) {
	placeErisConfig(`*`)
	defer removeErisDir()

	_, err := LoadViperConfig(configErisDir, "eris", "test")
	if err == nil {
		t.Fatalf("expected failure, got nil")
	}
}

func TestLoadViperConfigNonExistent1(t *testing.T) {
	_, err := LoadViperConfig(configErisDir, "eris", "test")
	if err == nil {
		t.Fatalf("expected failure, got nil")
	}
}

func TestLoadViperConfigNonExistent2(t *testing.T) {
	_, err := LoadViperConfig(configErisDir, "12345", "test")
	if err == nil {
		t.Fatalf("expected failure, got nil")
	}
}

func TestGetConfigValueUninitialized1(t *testing.T) {
	GlobalConfig = nil

	if returned := GetConfigValue("IpfsHost"); returned != "" {
		t.Fatalf("expected empty value, got %v", returned)
	}
}

func TestGetConfigValueUninitialized2(t *testing.T) {
	GlobalConfig = &ErisCli{}
	GlobalConfig.Config = nil

	if returned := GetConfigValue("IpfsHost"); returned != "" {
		t.Fatalf("expected empty value, got %v", returned)
	}
}

func TestGetConfigValue(t *testing.T) {
	var err error
	GlobalConfig, err = SetGlobalObject(os.Stdout, os.Stderr)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	if returned := GetConfigValue("IpfsHost"); returned == "" {
		t.Fatal("expected value returned, got empty string")
	}
	if returned := GetConfigValue("CompilersHost"); returned == "" {
		t.Fatal("expected value returned, got empty string")
	}
	if returned := GetConfigValue("DockerHost"); returned != "" {
		t.Fatalf("expected empty string returned, got %v", returned)
	}
	if returned := GetConfigValue("DockerCertPath"); returned != "" {
		t.Fatalf("expected empty string returned, got %v", returned)
	}
}

func TestGetConfigValueBad(t *testing.T) {
	var err error
	GlobalConfig, err = SetGlobalObject(os.Stdout, os.Stderr)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	if returned := GetConfigValue("bad value"); returned != "" {
		t.Fatalf("expected empty string returned, got %v", returned)
	}
}

func TestGitConfigUser(t *testing.T) {
	_, _, err := GitConfigUser()
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
}

func TestSaveGlobalConfig(t *testing.T) {
	os.MkdirAll(configErisDir, 0755)
	defer removeErisDir()

	config := &ErisConfig{
		IpfsHost:       "foo",
		CompilersHost:  "bar",
		DockerHost:     "baz",
		DockerCertPath: "qux",
		Verbose:        true,
	}
	if err := SaveGlobalConfig(config); err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	filename := filepath.Join(configErisDir, "eris.toml")
	expected := `IpfsHost = "foo"
CompilersHost = "bar"
DockerHost = "baz"
DockerCertPath = "qux"
Verbose = true
`
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Fatalf("expected config file created, it hasn't %v", err)
		t.FailNow()
	}

	if returned := fileContents(filename); returned != expected {
		t.Fatalf("expected certain file contents, got this %q", returned)
	}
}

func TestSaveGlobalConfigNotExistentDir(t *testing.T) {
	GlobalConfig = &ErisCli{}
	ChangeErisDir("/non/existent/dir")
	_, err := SetGlobalObject(os.Stderr, os.Stdout)
	if err != nil {
		t.Fatalf("expected success, got error %v", err)
	}

	config := &ErisConfig{
		IpfsHost:       "foo",
		CompilersHost:  "bar",
		DockerHost:     "baz",
		DockerCertPath: "qux",
		Verbose:        true,
	}
	if err := SaveGlobalConfig(config); err == nil {
		t.Fatal("expected failure, got nil")
	}
}

func TestSaveGlobalConfigNil(t *testing.T) {
	if err := SaveGlobalConfig(nil); err == nil {
		t.Fatal("expected failure, got nil")
	}
}

func TestChangeErisDirNonInitialized(t *testing.T) {
	GlobalConfig = nil
	ChangeErisDir(configErisDir)

	if GlobalConfig != nil {
		t.Fatal("didn't expect global config to become initialized")
	}
}

func TestChangeErisDir(t *testing.T) {
	GlobalConfig = &ErisCli{}
	ChangeErisDir(configErisDir)

	if GlobalConfig.ErisDir != configErisDir {
		t.Fatalf("expected config directory to change, got %v", GlobalConfig.ErisDir)
	}
}

func TestChangeErisDirCI(t *testing.T) {
	GlobalConfig = &ErisCli{}

	os.Setenv("TESTING", "true")
	ChangeErisDir(configErisDir)

	if GlobalConfig.ErisDir != "" {
		t.Fatalf("expected config directory not changed in CI, got %v", GlobalConfig.ErisDir)
	}
}

func placeErisConfig(definition string) {
	os.MkdirAll(configErisDir, 0755)
	fakeDefinitionFile(configErisDir, "eris", definition)
}

func removeErisDir() {
	// Move out of configErisDir before deleting it.
	parentPath := filepath.Join(configErisDir, "..")
	os.Chdir(parentPath)

	if err := os.RemoveAll(configErisDir); err != nil {
		panic(err)
	}
}

func fakeDefinitionFile(tmpDir, name, definition string) error {
	filename := filepath.Join(tmpDir, name+".toml")
	out, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = out.WriteString(definition)
	if err != nil {
		return err
	}

	return err
}

func fileContents(filename string) string {
	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	content, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}

	return string(content)
}
