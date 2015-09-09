package util

import (
	"os"
	"path/filepath"

	. "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/common"
)

func PathChecker(name string) (string, error) {
	pathS := filepath.Join(BlockchainsPath, "config", name)
	src, err := os.Stat(pathS)
	if err != nil || !src.IsDir() {
		logger.Infof("path: %s does not exist or is not a directory, please pass in a valid path or ensure a dir was created and has the correct files in it\n", pathS)
		return "", err
	}
	return pathS, nil
}
