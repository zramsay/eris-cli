package config

import (
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

var (
	// Convenience directories.
	GoPath             = os.Getenv("GOPATH")
	MonaxLtd           = filepath.Join(GoPath, "src", "github.com", "monax") // CSK: to deprecate
	MonaxGo            = filepath.Join(GoPath, "src", "github.com", "monax") // CSK: to keep
	MonaxGH            = "https://github.com/monax/"
	MonaxRoot          = ResolveMonaxRoot()
	MonaxContainerRoot = "/home/monax/.monax"

	// Major directories.
	AppsPath     = filepath.Join(MonaxRoot, "apps")
	BundlesPath  = filepath.Join(MonaxRoot, "bundles")
	ChainsPath   = filepath.Join(MonaxRoot, "chains")
	KeysPath     = filepath.Join(MonaxRoot, "keys")
	ScratchPath  = filepath.Join(MonaxRoot, "scratch")
	ServicesPath = filepath.Join(MonaxRoot, "services")

	// Chains directories.
	HEAD             = filepath.Join(ChainsPath, "HEAD")
	AccountsTypePath = filepath.Join(ChainsPath, "account-types")
	ChainTypePath    = filepath.Join(ChainsPath, "chain-types")

	// Keys directories.
	KeysDataPath      = filepath.Join(KeysPath, "data")
	KeysNamesPath     = filepath.Join(KeysPath, "names")
	KeysContainerPath = path.Join(MonaxContainerRoot, "keys", "data")

	// Scratch directories.
	DataContainersPath   = filepath.Join(ScratchPath, "data")
	LanguagesScratchPath = filepath.Join(ScratchPath, "languages")
	LllcScratchPath      = filepath.Join(LanguagesScratchPath, "lllc")
	SolcScratchPath      = filepath.Join(LanguagesScratchPath, "sol")
	SerpScratchPath      = filepath.Join(LanguagesScratchPath, "ser")
)

// DirsToMigrate is used by the `monax init` command to check
// if old dirs exist to migrate them.
var DirsToMigrate = map[string]string{}

func HomeDir() string {
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

// ChangeMonaxRoot points the root of the Monax settings hierarchy
// to the monaxDir location.
func ChangeMonaxRoot(monaxDir string) {
	if os.Getenv("TESTING") == "true" {
		return
	}

	MonaxRoot = monaxDir

	// Major directories.
	AppsPath = filepath.Join(MonaxRoot, "apps")     // previously "dapps"
	ChainsPath = filepath.Join(MonaxRoot, "chains") // previously "blockchains"
	KeysPath = filepath.Join(MonaxRoot, "keys")
	ScratchPath = filepath.Join(MonaxRoot, "scratch")
	ServicesPath = filepath.Join(MonaxRoot, "services")

	// Chains Directories
	AccountsTypePath = filepath.Join(ChainsPath, "account-types")
	ChainTypePath = filepath.Join(ChainsPath, "chain-types")
	HEAD = filepath.Join(ChainsPath, "HEAD")

	// Keys Directories
	KeysDataPath = filepath.Join(KeysPath, "data")
	KeysNamesPath = filepath.Join(KeysPath, "names")

	// Scratch Directories (basically monax' cache) (globally coordinated)
	DataContainersPath = filepath.Join(ScratchPath, "data")
	LanguagesScratchPath = filepath.Join(ScratchPath, "languages") // previously "~/.monax/languages"
}

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

// TODO: [csk] give this a default string if folks want it somewhere besides ~/.monax ...?
func ResolveMonaxRoot() string {
	var monax string
	if os.Getenv("MONAX") != "" {
		monax = os.Getenv("MONAX")
	} else {
		if runtime.GOOS == "windows" {
			home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
			if home == "" {
				home = os.Getenv("USERPROFILE")
			}
			monax = filepath.Join(home, ".monax")
		} else {
			monax = filepath.Join(HomeDir(), ".monax")
		}
	}
	return monax
}

// InitMonaxDir creates an Monax directory hierarchy under MonaxRoot dir.
func InitMonaxDir() (err error) {
	for _, d := range []string{
		MonaxRoot,
		AppsPath,
		BundlesPath,
		ChainsPath,
		AccountsTypePath,
		ChainTypePath,
		KeysPath,
		KeysDataPath,
		KeysNamesPath,
		ScratchPath,
		DataContainersPath,
		LanguagesScratchPath,
		LllcScratchPath,
		SolcScratchPath,
		SerpScratchPath,
		ServicesPath,
	} {
		err := InitDataDir(d)
		if err != nil {
			return err
		}
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
		tmpDir, err := ioutil.TempDir(os.TempDir(), "monax_copy")
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
	if err != nil {
		return err
	}
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
	return err
}

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
