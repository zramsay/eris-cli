package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/monax/cli/log"
)

var (
	configMonaxDir = filepath.Join(os.TempDir(), "config")
)

func TestMain(m *testing.M) {
	log.SetLevel(log.ErrorLevel)
	// log.SetLevel(log.InfoLevel)
	// log.SetLevel(log.DebugLevel)

	// Unset this variable by default for config package.
	savedEnv := os.Getenv("TESTING")
	if err := os.Unsetenv("TESTING"); err != nil {
		panic("can't unset TESTING")
	}
	defer os.Setenv("TESTING", savedEnv)

	log.WithField("dir", configMonaxDir).Info("Using temporary directory for config files")

	m.Run()
}

func TestNew1(t *testing.T) {
	cli, err := New(os.Stdout, os.Stderr)
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

func TestNew2(t *testing.T) {
	cli, err := New(os.Stderr, os.Stdout)
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

func TestNewNil(t *testing.T) {
	cli, err := New(nil, nil)
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

func TestNewDefaultConfig(t *testing.T) {
	ChangeMonaxRoot(configMonaxDir)

	cli, err := New(os.Stderr, os.Stdout)
	if err != nil {
		t.Fatalf("expected success, got error %v", err)
	}

	defaults, err := SetDefaults()
	if err != nil {
		t.Fatalf("expected defaults loaded, got error %v", err)
	}

	if def, returned := defaults.Get("IpfsHost"), cli.IpfsHost; reflect.DeepEqual(returned, def) != true {
		t.Fatalf("expected default %q, got %q", returned, def)
	}

	if def, returned := defaults.Get("CompilersHost"), cli.CompilersHost; reflect.DeepEqual(returned, def) != true {
		t.Fatalf("expected default %q, got %q", returned, def)
	}

	log.WithFields(log.Fields{
		"ipfshost":       cli.IpfsHost,
		"compilers host": cli.CompilersHost,
		"host":           cli.DockerHost,
		"cert path":      cli.DockerCertPath,
		"crash report":   cli.CrashReport,
		"verbose":        cli.Verbose,
	}).Info("Checking defaults")
}

func TestNewCustomConfig(t *testing.T) {
	placeSettings(`
IpfsHost = "foo"
CompilersHost = "bar"
DockerHost = "baz"
DockerCertPath = "qux"
CrashReport = "quux"
Verbose = true
`)
	defer removeMonaxDir()

	ChangeMonaxRoot(configMonaxDir)
	cli, err := New(os.Stderr, os.Stdout)
	if err != nil {
		t.Fatalf("expected success, got error %v", err)
	}

	if custom, returned := "foo", cli.IpfsHost; custom != returned {
		t.Fatalf("expected %q, got %q", custom, returned)
	}
	if custom, returned := "bar", cli.CompilersHost; custom != returned {
		t.Fatalf("expected %q, got %q", custom, returned)
	}
	if custom, returned := "baz", cli.DockerHost; custom != returned {
		t.Fatalf("expected %q, got %q", custom, returned)
	}
	if custom, returned := "qux", cli.DockerCertPath; custom != returned {
		t.Fatalf("expected %q, got %q", custom, returned)
	}
	if custom, returned := "quux", cli.CrashReport; custom != returned {
		t.Fatalf("expected %q, got %q", custom, returned)
	}
	if custom, returned := true, cli.Verbose; custom != returned {
		t.Fatalf("expected %v, got %v", custom, returned)
	}
}

func TestNewCustomEmptyConfig(t *testing.T) {
	placeSettings(``)
	defer removeMonaxDir()

	ChangeMonaxRoot(configMonaxDir)
	cli, err := New(os.Stderr, os.Stdout)
	if err != nil {
		t.Fatalf("expected success, got error %v", err)
	}

	defaults, err := SetDefaults()
	if err != nil {
		t.Fatalf("expected defaults loaded, got error %v", err)
	}

	log.WithFields(log.Fields{
		"ipfs host":      cli.IpfsHost,
		"compilers host": cli.CompilersHost,
		"host":           cli.DockerHost,
		"cert path":      cli.DockerCertPath,
		"crash report":   cli.CrashReport,
		"verbose":        cli.Verbose,
	}).Info("Checking empty values")

	// With an empty config, the values are used are defaults.
	if def, returned := defaults.Get("IpfsHost"), cli.IpfsHost; reflect.DeepEqual(returned, def) != true {
		t.Fatalf("expected default %v, got %v", returned, def)
	}

	if def, returned := defaults.Get("CompilersHost"), cli.CompilersHost; reflect.DeepEqual(returned, def) != true {
		t.Fatalf("expected default %q, got %q", returned, def)
	}

	if custom, returned := "", cli.DockerHost; custom != returned {
		t.Fatalf("expected %q, got %q", custom, returned)
	}
	if custom, returned := "", cli.DockerCertPath; custom != returned {
		t.Fatalf("expected %q, got %q", custom, returned)
	}
	if custom, returned := "bugsnag", cli.CrashReport; custom != returned {
		t.Fatalf("expected %q, got %q", custom, returned)
	}
	if custom, returned := false, cli.Verbose; custom != returned {
		t.Fatalf("expected %v, got %v", custom, returned)
	}
}

func TestNewCustomBadConfig(t *testing.T) {
	placeSettings(`*`)
	defer removeMonaxDir()

	ChangeMonaxRoot(configMonaxDir)
	cli, err := New(os.Stderr, os.Stdout)
	if err != nil {
		t.Fatalf("expected success, got error %v", err)
	}

	log.WithFields(log.Fields{
		"ipfshost":       cli.IpfsHost,
		"compilers host": cli.CompilersHost,
		"host":           cli.DockerHost,
		"cert path":      cli.DockerCertPath,
		"crash report":   cli.CrashReport,
		"verbose":        cli.Verbose,
	}).Info("Checking empty values")

	// With an empty config, the values are used are defaults.
	defaults, err := SetDefaults()
	if err != nil {
		t.Fatalf("expected defaults loaded, got error %v", err)
	}

	if def, returned := defaults.Get("IpfsHost"), cli.IpfsHost; reflect.DeepEqual(returned, def) != true {
		t.Fatalf("expected default %q, got %q", returned, def)
	}

	if def, returned := defaults.Get("CompilersHost"), cli.CompilersHost; reflect.DeepEqual(returned, def) != true {
		t.Fatalf("expected default %q, got %q", returned, def)
	}

	if custom, returned := "", cli.DockerHost; custom != returned {
		t.Fatalf("expected %q, got %q", custom, returned)
	}
	if custom, returned := "", cli.DockerCertPath; custom != returned {
		t.Fatalf("expected %q, got %q", custom, returned)
	}
	if custom, returned := "bugsnag", cli.CrashReport; custom != returned {
		t.Fatalf("expected %q, got %q", custom, returned)
	}
	if custom, returned := false, cli.Verbose; custom != returned {
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

func TestLoad(t *testing.T) {
	placeSettings(`
IpfsHost = "foo"
CompilersHost = "bar"
DockerHost = "baz"
DockerCertPath = "qux"
CrashReport = "quux"
Verbose = true
`)
	defer removeMonaxDir()

	config, err := Load()
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

func TestLoadEmpty(t *testing.T) {
	placeSettings(``)
	defer removeMonaxDir()

	config, err := Load()
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

func TestLoadBad(t *testing.T) {
	placeSettings(`*`)
	defer removeMonaxDir()

	// With bad config, load defaults.
	config, err := Load()
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

func TestLoadViper(t *testing.T) {
	placeSettings(`
IpfsHost = "foo"
CompilersHost = "bar"
DockerHost = "baz"
DockerCertPath = "qux"
CrashReport = "quux"
Verbose = true
`)
	defer removeMonaxDir()

	config, err := LoadViper(configMonaxDir, "monax")
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

func TestLoadViperEmpty(t *testing.T) {
	placeSettings(``)
	defer removeMonaxDir()

	config, err := LoadViper(configMonaxDir, "monax")
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

func TestLoadViperBad(t *testing.T) {
	placeSettings(`*`)
	defer removeMonaxDir()

	_, err := LoadViper(configMonaxDir, "monax")
	if err == nil {
		t.Fatalf("expected failure, got nil")
	}
}

func TestLoadViperNonExistent1(t *testing.T) {
	_, err := LoadViper(configMonaxDir, "monax")
	if err == nil {
		t.Fatalf("expected failure, got nil")
	}
}

func TestLoadViperNonExistent2(t *testing.T) {
	_, err := LoadViper(configMonaxDir, "12345")
	if err == nil {
		t.Fatalf("expected failure, got nil")
	}
}

func TestSave(t *testing.T) {
	os.MkdirAll(configMonaxDir, 0755)
	defer removeMonaxDir()

	settings := &Settings{
		IpfsHost:       "foo",
		CompilersHost:  "bar",
		DockerHost:     "baz",
		DockerCertPath: "qux",
		Verbose:        true,
	}
	if err := Save(settings); err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	filename := filepath.Join(configMonaxDir, "monax.toml")
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

func TestSaveNotExistentDir(t *testing.T) {
	ChangeMonaxRoot("/non/existent/dir")
	_, err := New(os.Stderr, os.Stdout)
	if err != nil {
		t.Fatalf("expected success, got error %v", err)
	}

	settings := &Settings{
		IpfsHost:       "foo",
		CompilersHost:  "bar",
		DockerHost:     "baz",
		DockerCertPath: "qux",
		Verbose:        true,
	}
	if err := Save(settings); err == nil {
		t.Fatal("expected failure, got nil")
	}
}

func TestSaveNil(t *testing.T) {
	if err := Save(nil); err == nil {
		t.Fatal("expected failure, got nil")
	}
}

func placeSettings(definition string) {
	os.MkdirAll(configMonaxDir, 0755)
	fakeDefinitionFile(configMonaxDir, "monax", definition)
}

func removeMonaxDir() {
	// Move out of configMonaxDir before deleting it.
	parentPath := filepath.Join(configMonaxDir, "..")
	os.Chdir(parentPath)

	if err := os.RemoveAll(configMonaxDir); err != nil {
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
