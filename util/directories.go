package util

import (
	"os"
	"path"
	"runtime"
)

func UserErisDir() string {
	var eris string
	if (os.Getenv("ERISROOT") != "") {
		eris = os.Getenv("ERISROOT")
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
