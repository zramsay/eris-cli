package util

import (
	"os"
	"path/filepath"

	. "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/common"
)

func ChainsPathChecker(name string) (string, error) {
	pathS := filepath.Join(BlockchainsPath, name)
	src, err := os.Stat(pathS)
	if err != nil || !src.IsDir() {
		logger.Infof("path: %s does not exist or is not a directory, please pass in a valid path or ensure a dir was created and has the correct files in it\n", pathS)
		return "", err
	}
	return pathS, nil
}

func MoveOutOfDirAndRmDir(src, dest string) error {
	logger.Infof("Move all files/dirs out of a dir and rm -rf that dir.\n")
	logger.Debugf("Source of the move =>\t\t%s.\n", src)
	logger.Debugf("Destin of the move =>\t\t%s.\n", dest)
	toMove, err := filepath.Glob(filepath.Join(src, "*"))
	if err != nil {
		return err
	}

	if len(toMove) == 0 {
		logger.Debugln("No files to move.")
	}

	for _, f := range toMove {
		logger.Debugf("Moving [%s] to [%s].\n", f, filepath.Join(dest, filepath.Base(f)))

		// using a copy (read+write) strategy to get around swap partitions and other
		//   problems that cause a simple rename strategy to fail. it is more io overhead
		//   to do this, but for now that is preferable to alternative solutions.
		Copy(f, filepath.Join(dest, filepath.Base(f)))
	}

	logger.Infof("Removing directory =>\t\t%s.\n", src)
	err = os.RemoveAll(src)
	if err != nil {
		return err
	}

	return nil
}
