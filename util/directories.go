package util

import (
	"os"
	"path"
	"runtime"
)

func UserErisDir() string {
	var eris string
	if os.Getenv("ERIS_DIR") != "" {
		eris = os.Getenv("ERIS_DIR")
	} else {
		if runtime.GOOS == "windows" {
			home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
			if home == "" {
				home = os.Getenv("USERPROFILE")
			}
			eris = path.Join(home, ".eris")
		} else {
			eris = path.Join(os.Getenv("HOME"), ".eris")
		}
	}
	return eris
}

var (
	GoPath          = os.Getenv("GOPATH")
	ErisGoPath      = path.Join(GoPath, "src", "github.com", "eris-ltd")
	ErisRoot        = UserErisDir()
	ActionsPath     = path.Join(ErisRoot, "actions")
	KeysPath        = path.Join(ErisRoot, "keys")
	ServicesPath    = path.Join(ErisRoot, "services")
	BlockchainsPath = path.Join(ErisRoot, "blockchains")
	DappsPath       = path.Join(ErisRoot, "dapps")
	FilesystemsPath = path.Join(ErisRoot, "files")
	LanguagesPath   = path.Join(ErisRoot, "languages")
	LogsPath        = path.Join(ErisRoot, "logs")
	ScratchPath     = path.Join(ErisRoot, "scratch")
	HEAD            = path.Join(BlockchainsPath, "HEAD")
	RefsPath        = path.Join(BlockchainsPath, "refs")
	EpmScratch      = path.Join(ScratchPath, "epm")
	LllcScratch     = path.Join(ScratchPath, "lllc")
)

var MajorDirs = []string{
	ErisRoot, ActionsPath, KeysPath, DappsPath, BlockchainsPath, FilesystemsPath, LanguagesPath, LogsPath, ServicesPath, ScratchPath, RefsPath, EpmScratch, LllcScratch,
}
