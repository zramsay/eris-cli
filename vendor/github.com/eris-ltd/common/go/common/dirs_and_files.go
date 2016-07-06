package common

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

var (
	// Convenience Directories
	GoPath            = os.Getenv("GOPATH")
	ErisLtd           = filepath.Join(GoPath, "src", "github.com", "eris-ltd") // CSK: to deprecate
	ErisGo            = filepath.Join(GoPath, "src", "github.com", "eris-ltd") // CSK: to keep
	ErisGH            = "https://github.com/eris-ltd/"
	ErisRoot          = ResolveErisRoot()
	ErisContainerRoot = "/home/eris/.eris" // XXX: this is used as root in the `eris/base` image

	// Major Directories
	ActionsPath  = filepath.Join(ErisRoot, "actions")
	AppsPath     = filepath.Join(ErisRoot, "apps") // previously "dapps"
	BundlesPath  = filepath.Join(ErisRoot, "bundles")
	ChainsPath   = filepath.Join(ErisRoot, "chains") // previously "blockchains"
	KeysPath     = filepath.Join(ErisRoot, "keys")
	RemotesPath  = filepath.Join(ErisRoot, "remotes")
	ScratchPath  = filepath.Join(ErisRoot, "scratch")
	ServicesPath = filepath.Join(ErisRoot, "services")

	// Chains Directories
	HEAD             = filepath.Join(ChainsPath, "HEAD")
	DefaultChainPath = filepath.Join(ChainsPath, "default")
	AccountsTypePath = filepath.Join(ChainsPath, "account-types")
	ChainTypePath    = filepath.Join(ChainsPath, "chain-types")

	// Keys Directories
	KeysDataPath = filepath.Join(KeysPath, "data")
	KeyNamesPath = filepath.Join(KeysPath, "names")

	// Scratch Directories (basically eris' cache) (globally coordinated)
	DataContainersPath   = filepath.Join(ScratchPath, "data")
	LanguagesScratchPath = filepath.Join(ScratchPath, "languages") // previously "~/.eris/languages"
	LllcScratchPath      = filepath.Join(LanguagesScratchPath, "lllc")
	SolcScratchPath      = filepath.Join(LanguagesScratchPath, "sol")
	SerpScratchPath      = filepath.Join(LanguagesScratchPath, "ser")

	// Services Directories
	PersonalServicesPath = filepath.Join(ServicesPath, "global")

	// Deprecated Directories (remove on 0.12 release)
	BlockchainsPath = filepath.Join(ErisRoot, "blockchains")
	DappsPath       = filepath.Join(ErisRoot, "dapps")
	LanguagesPath   = filepath.Join(ErisRoot, "languages")
)

var MajorDirs = []string{
	ErisRoot,
	ActionsPath,
	AppsPath,
	BundlesPath,
	ChainsPath,
	DefaultChainPath,
	AccountsTypePath,
	ChainTypePath,
	KeysPath,
	KeysDataPath,
	KeyNamesPath,
	RemotesPath,
	ScratchPath,
	DataContainersPath,
	LanguagesScratchPath,
	LllcScratchPath,
	SolcScratchPath,
	SerpScratchPath,
	ServicesPath,
	PersonalServicesPath,
}

// These should only be used by specific tooling rather than eris-cli level
var ChainsDirs = []string{
	ChainsPath,
	DefaultChainPath,
	AccountsTypePath,
	ChainTypePath,
}

// These should only be used by specific tooling rather than eris-cli level
var KeysDirs = []string{
	KeysPath,
	KeysDataPath,
	KeyNamesPath,
}

// These should only be used by specific tooling rather than eris-cli level
var ServicesDirs = []string{
	ServicesPath,
	PersonalServicesPath,
}

// These should only be used by specific tooling rather than eris-cli level
var ScratchDirs = []string{
	ScratchPath,
	DataContainersPath,
	LanguagesScratchPath,
	LllcScratchPath,
	SolcScratchPath,
	SerpScratchPath,
}

//eris update checks if old dirs exist & migrates them
var DirsToMigrate = map[string]string{
	BlockchainsPath: ChainsPath,
	DappsPath:       AppsPath,
	LanguagesPath:   LanguagesScratchPath,
}

//---------------------------------------------
// user and process

func Usr() string {
	if runtime.GOOS == "windows" {
		drive := os.Getenv("HOMEDRIVE")
		path := os.Getenv("HOMEPATH")
		if drive == "" || path == "" {
			return os.Getenv("USERPROFILE")
		}
		return drive + path
	} else {
		return os.Getenv("HOME")
	}
}

func Exit(err error) {
	status := 0
	if err != nil {
		fmt.Println(err)
		status = 1
	}
	os.Exit(status)
}

func IfExit(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// user and process
//---------------------------------------------------------------------------
// filesystem

func AbsolutePath(Datadir string, filename string) string {
	if filepath.IsAbs(filename) {
		return filename
	}
	return filepath.Join(Datadir, filename)
}

func InitDataDir(Datadir string) error {
	if _, err := os.Stat(Datadir); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(Datadir, 0777); err != nil {
				return err
			}
		}
	}
	return nil
}

// TODO: [csk] give this a default string if folks want it somewhere besides ~/.eris ...?
func ResolveErisRoot() string {
	var eris string
	if os.Getenv("ERIS") != "" {
		eris = os.Getenv("ERIS")
	} else {
		if runtime.GOOS == "windows" {
			home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
			if home == "" {
				home = os.Getenv("USERPROFILE")
			}
			eris = filepath.Join(home, ".eris")
		} else {
			eris = filepath.Join(Usr(), ".eris")
		}
	}
	return eris
}

// Create the default eris tree
func InitErisDir() (err error) {
	for _, d := range MajorDirs {
		err := InitDataDir(d)
		if err != nil {
			return err
		}
	}
	if _, err = os.Stat(HEAD); err != nil {
		_, err = os.Create(HEAD)
	}
	return
}

func ClearDir(dir string) error {
	fs, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, f := range fs {
		n := f.Name()
		if f.IsDir() {
			if err := os.RemoveAll(filepath.Join(dir, f.Name())); err != nil {
				return err
			}
		} else {
			if err := os.Remove(filepath.Join(dir, n)); err != nil {
				return err
			}
		}
	}
	return nil
}

func Copy(src, dst string) error {
	f, err := os.Stat(src)
	if err != nil {
		return err
	}
	if f.IsDir() {
		tmpDir, err := ioutil.TempDir(os.TempDir(), "eris_copy")
		if err != nil {
			return err
		}
		if err := copyDir(src, tmpDir); err != nil {
			return err
		}
		if err := copyDir(tmpDir, dst); err != nil {
			return err
		}
		return nil
	}
	return copyFile(src, dst)
}

// assumes we've done our checking
func copyDir(src, dst string) error {
	fi, err := os.Stat(src)
	if err := os.MkdirAll(dst, fi.Mode()); err != nil {
		return err
	}
	fs, err := ioutil.ReadDir(src)
	if err != nil {
		return err
	}

	for _, f := range fs {
		s := filepath.Join(src, f.Name())
		d := filepath.Join(dst, f.Name())
		if f.IsDir() {
			if err := copyDir(s, d); err != nil {
				return err
			}
		} else {
			if err := copyFile(s, d); err != nil {
				return err
			}
		}
	}
	return nil
}

// common golang, really?
func copyFile(src, dst string) error {
	r, err := os.Open(src)
	if err != nil {
		return err
	}
	defer r.Close()

	w, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer w.Close()

	_, err = io.Copy(w, r)
	if err != nil {
		return err
	}
	return nil
}

func WriteFile(data, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0775); err != nil {
		return err
	}
	writer, err := os.Create(filepath.Join(path))
	defer writer.Close()
	if err != nil {
		return err
	}
	writer.Write([]byte(data))
	return nil
}

// filesystem
//-------------------------------------------------------
// open text editors

func Editor(file string) error {
	editr := os.Getenv("EDITOR")
	if strings.Contains(editr, "/") {
		editr = filepath.Base(editr)
	}
	switch editr {
	case "", "vim", "vi":
		return vi(file)
	case "emacs":
		return emacs(file)
	default:
		return editor(file)
	}
}

func emacs(file string) error {
	cmd := exec.Command("emacs", file)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func vi(file string) error {
	cmd := exec.Command("vim", file)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func editor(file string) error {
	cmd := exec.Command(os.Getenv("EDITOR"), file)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
