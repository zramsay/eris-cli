package util

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/monax/monax/log"
)

//XXX this command absolutely needs a good test!!
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

	return fmt.Errorf("permission to migrate not given")
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
			return fmt.Errorf("neither deprecated (%s) or new (%s) exists. please run `init` prior to `update`\n", depDir, newDir)
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
		} else { //should never throw
			return fmt.Errorf("unknown and unresolveable conflict between directory to deprecate (%s) and new directory (%s)\n", depDir, newDir)
		}
		if DoesDirExist(depDir) {
			return fmt.Errorf("deprecated directory (%s) still exists, something went wrong", depDir)
		}
	}

	return nil
}

func checkFileNamesAndMigrate(depDir, newDir string) error {
	depDirFiles, err := ioutil.ReadDir(depDir)
	if err != nil {
		return fmt.Errorf("could not read files from dir to be deprecated %s:\n%v\n", depDir, err)
	}
	newDirFiles, err := ioutil.ReadDir(newDir)
	if err != nil {
		return fmt.Errorf("could not read files from new dir %s:\n%v\n", newDir, err)
	}

	fileNamesToCheck := make(map[string]bool) // map of filenames in new dir
	if len(newDirFiles) != 0 {
		for _, file := range newDirFiles {
			fileNamesToCheck[file.Name()] = true
		}
	}

	for _, file := range depDirFiles { // if any filenames match, must resolve
		depFile := filepath.Join(depDir, file.Name())
		newFile := filepath.Join(newDir, file.Name()) // file may not actually exist (yet)

		if fileNamesToCheck[file.Name()] { // conflict!
			oldFileContents, _ := ioutil.ReadFile(depFile)
			newFileContents, _ := ioutil.ReadFile(newFile)
			if string(newFileContents) != string(oldFileContents) {
				return fmt.Errorf("identical filename; different content; identified in deprecated dir (%s) and new dir to migrate to (%s)\nplease resolve and re-run command", depFile, newFile)
			} else { // same file so no need to move
				continue
			}
		} else { // filenames don't match, move file from depDir to newDir
			if err := os.Rename(depFile, newFile); err != nil {
				log.WithFields(log.Fields{
					"from": depFile,
					"to":   newFile,
				}).Warn("File migration NOT successful")
				return err
			}
			log.WithFields(log.Fields{
				"from": depFile,
				"to":   newFile,
			}).Warn("File migration successful")
		}
	}
	return nil
}
