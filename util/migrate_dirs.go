package util

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	log "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/Sirupsen/logrus"
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
	} else {
		return fmt.Errorf("permission to migrate not given")
	}

	return nil
}

//check that migration is actually needed
func dirCheckMaker(dirsToMigrate map[string]string) (map[string]string, bool) {

	for depDir, newDir := range dirsToMigrate {
		if !DoesDirExist(depDir) && DoesDirExist(newDir) { //already migrated, nothing to see here
			return nil, false
		} else {
			return dirsToMigrate, true
		}
	}
	return dirsToMigrate, true
}

func canWeMigrate() bool {
	fmt.Print("Permission to migrate deprecated directories required: would you like to continue? (Y/y): ")
	var input string
	fmt.Scanln(&input)
	if input == "Y" || input == "y" || input == "YES" || input == "Yes" || input == "yes" {
		log.Debug("Confirmation verified. Proceeding")
		return true
	} else {
		return false
	}
}

func Migrate(dirsToMigrate map[string]string) error {
	for depDir, newDir := range dirsToMigrate {
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
			// [csk] once the files are migrated we need to remove the dir or
			// the DoesDirExist function will return.
			if err := os.Remove(depDir); err != nil {
				return err
			}
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

	fileNamesToCheck := make(map[string]bool) //map of filenames in new dir
	if len(newDirFiles) != 0 {
		for _, file := range newDirFiles {
			fileNamesToCheck[file.Name()] = true
		}
	}

	for _, file := range depDirFiles { //if any filenames match, must resolve
		depFile := filepath.Join(depDir, file.Name())
		newFile := filepath.Join(newDir, file.Name()) //file may not actually exist (yet)

		if fileNamesToCheck[file.Name()] == true { //conflict!
			return fmt.Errorf("identical file name in deprecated dir (%s) and new dir to migrate to (%s)\nplease resolve and re-run command", depFile, newFile)
		} else { //filenames don't match, move file from depDir to newDir
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

func DoesDirExist(dir string) bool {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return false
	} else {
		return true
	}
}
