package util

import (
	"os"
	"path/filepath"
	"strings"

	log "github.com/eris-ltd/eris-logger"

	. "github.com/eris-ltd/common/go/common"
)

func ChainsPathChecker(name string) (string, error) {
	pathS := filepath.Join(ChainsPath, name)
	src, err := os.Stat(pathS)
	if err != nil || !src.IsDir() {
		log.WithField("=>", pathS).Info("Path does not exist or not a diretory")
		return "", err
	}
	return pathS, nil
}

func GetFileByNameAndType(typ, name string) string {
	log.WithFields(log.Fields{
		"file": name,
		"type": typ,
	}).Debug("Looking for file")
	files := GetGlobalLevelConfigFilesByType(typ, true)

	for _, file := range files {
		fileBase := strings.Split(filepath.Base(file), ".")[0] // quick and dirty file root
		if fileBase == name {
			log.WithField("file", file).Debug("This file found")
			return file
		}
		log.WithField("file", file).Debug("Group file found")
	}

	return ""
}

// note this function fails silently.
func GetGlobalLevelConfigFilesByType(typ string, withExt bool) []string {
	var path string
	switch typ {
	case "services":
		path = ServicesPath
	case "chains":
		path = ChainsPath
	case "actions":
		path = ActionsPath
	}

	files := []string{}
	fileTypes := []string{}

	// TODO [csk]: DRY up how we deal with file extensions
	for _, t := range []string{"*.json", "*.yaml", "*.toml"} {
		fileTypes = append(fileTypes, filepath.Join(path, t))
	}

	for _, t := range fileTypes {
		s, _ := filepath.Glob(t)
		for _, s1 := range s {
			if !withExt {
				s1 = strings.Split(filepath.Base(s1), ".")[0]
			}
			files = append(files, s1)
		}
	}
	return files
}

func MoveOutOfDirAndRmDir(src, dest string) error {
	log.WithFields(log.Fields{
		"from": src,
		"to":   dest,
	}).Info("Move all files/dirs out of a dir and `rm -fr` that dir")
	toMove, err := filepath.Glob(filepath.Join(src, "*"))
	if err != nil {
		return err
	}

	if len(toMove) == 0 {
		log.Debug("No files to move")
	}

	for _, f := range toMove {
		t := filepath.Join(dest, filepath.Base(f))
		log.WithFields(log.Fields{
			"from": f,
			"to":   t,
		}).Debug("Moving")

		// using a copy (read+write) strategy to get around swap partitions and other
		//   problems that cause a simple rename strategy to fail. it is more io overhead
		//   to do this, but for now that is preferable to alternative solutions.
		Copy(f, t)
	}

	log.WithField("=>", src).Info("Removing directory")
	err = os.RemoveAll(src)
	if err != nil {
		return err
	}

	return nil
}
