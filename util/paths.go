package util

import (
	"os"
	"path/filepath"

	. "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/common"
)

func PathChecker(typ, name string) (string, error) {
	var pathS string
	if typ == "chain" {
		pathS = filepath.Join(BlockchainsPath, "config", name)
	}
	src, err := os.Stat(pathS)
	if err != nil {
		logger.Printf("path: %s does not exist, please pass in a valid path\n", pathS)
		return "", err
	}
	if !src.IsDir() {
		logger.Errorf("path: %s is not a directory, please ensure a dir was created and has the correct files in it\n", pathS)
		return "", err

	}
	return pathS, nil
}
