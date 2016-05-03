package util

import (
	"io/ioutil"
	"os"
	"path/filepath"

	log "github.com/eris-ltd/eris-logger"
	. "github.com/eris-ltd/eris-cli/errors"

)

func MigrateDeprecatedDirs(dirsToMigrate map[string]string, prompt bool) error {
	dirsMap, isMigNeed := dirCheckMaker(dirsToMigrate)
	if isMigNeed {
		log.Warn("Deprecated directories detected. Marmot migration commencing")
	}

	if !isMigNeed {
		log.Info("Nothing to migrate")
		return nil
	} else if !prompt {
		return Migrate(dirsMap)
	} else if canWeMigrate() {
		return Migrate(dirsMap)
	}

	return ErrNoPermGiven
}

//check that migration is actually needed
func dirCheckMaker(dirsToMigrate map[string]string) (map[string]string, bool) {
	newMigration := make(map[string]string)

	for depDir, newDir := range dirsToMigrate {
		log.WithFields(log.Fields{
			"old":        depDir,
			"old exists": DoesDirExist(depDir),
			"new":        newDir,
			"new exists": DoesDirExist(newDir),
		}).Debug("Checking Directories to Migrate")
		if !DoesDirExist(depDir) && DoesDirExist(newDir) { //already migrated, nothing to see here
			continue
		} else {
			newMigration[depDir] = newDir
		}
	}
	return newMigration, (len(newMigration) > 0)
}

func canWeMigrate() bool {
	log.Warn("Permission to migrate deprecated directories required")
	if QueryYesOrNo("Would you like to continue?") == Yes {
		log.Debug("Confirmation verified. Proceeding")
		return true
	} else {
		return false
	}
}

func Migrate(dirsToMigrate map[string]string) error {
	for depDir, newDir := range dirsToMigrate {
		log.WithFields(log.Fields{
			"old": depDir,
			"new": newDir,
		}).Info("Migrating Directories")
		if !DoesDirExist(depDir) && !DoesDirExist(newDir) {
			return BaseErrorESS(ErrNoDirectories, depDir, newDir)
		} else if DoesDirExist(depDir) && !DoesDirExist(newDir) { //never updated, just rename dirs
			if err := os.Rename(depDir, newDir); err != nil {
				return err
			}
			log.WithFields(log.Fields{
				"from": depDir,
				"to":   newDir,
			}).Warn("Directory migration successful")
		} else if DoesDirExist(depDir) && DoesDirExist(newDir) { //both exist, better check what's in them
			if err := checkFileNamesAndMigrate(depDir, newDir); err != nil {
				return err
			}
			if err := os.Remove(depDir); err != nil {
				return err
			}
		} else if !DoesDirExist(depDir) && DoesDirExist(newDir) { // old is gone, new is there; continue (dirCheckMaker should check this)
			continue
		}
	}

	return nil
}

func checkFileNamesAndMigrate(depDir, newDir string) error {
	depDirFiles, err := ioutil.ReadDir(depDir)
	if err != nil {
		return err
	}
	newDirFiles, err := ioutil.ReadDir(newDir)
	if err != nil {
		return err
	}

	fileNamesToCheck := make(map[string]bool) //map of filenames in new dir
	if len(newDirFiles) != 0 {
		for _, file := range newDirFiles {
			fileNamesToCheck[file.Name()] = true
		}
	}

	for _, file := range depDirFiles { //if any filenames match, must resolve
		depFile := filepath.Join(depDir, file.Name())
		newFile := filepath.Join(newDir, file.Name()) //file may not actually exist (yet)

		if err := os.Rename(depFile, newFile); err != nil {
			log.WithFields(log.Fields{
				"from": depFile,
				"to":   newFile,
			}).Warn("File migration not successful")
			return err
		}

		log.WithFields(log.Fields{
			"from": depFile,
			"to":   newFile,
		}).Warn("File migration successful")
	}
	return nil
}

func DoesDirExist(dir string) bool {
	f, err := os.Stat(dir)
	if err != nil {
		return false
	}
	if !f.IsDir() {
		return false
	}
	return true
}
