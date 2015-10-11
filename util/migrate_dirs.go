package util

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
)

//XXX this command absolutely needs a good test!!
func MigrateDeprecatedDirs(depDirs, newDirs []string, prompt bool) error {
	if len(depDirs) != len(newDirs) {
		return fmt.Errorf("Number of dirs to deprecate (%d) does not match # of new dirs (%d)\n", depDirs, newDirs)
	}

	dirsMap, isMigNeed := dirCheckMaker(depDirs, newDirs)
	if isMigNeed {
		logger.Println("deprecated directories detected, marmot migration commencing")
	}

	if !isMigNeed {
		return nil
	} else if !prompt {
		Migrate(dirsMap)
	} else if canWeMigrate() {
		Migrate(dirsMap)
	} else {
		return fmt.Errorf("permission to migrate not given")
	}

	return nil
}

//check that migration is actually needed and make map
func dirCheckMaker(depDirs, newDirs []string) (map[string]string, bool) {
	dirsToMigrate := make(map[string]string)
	for i, d := range depDirs {
		dirsToMigrate[d] = newDirs[i]
	}

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
	logger.Printf("permission to migrate deprecated directories required: would you like to continue? (Y/y)\n")
	var input string
	fmt.Scanln(&input)
	if input == "Y" || input == "y" || input == "YES" || input == "Yes" || input == "yes" {
		logger.Debugf("Confirmation verified. Proceeding.\n")
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
			logger.Printf("Directory migration succesful:\t%s ====> %s\n", depDir, newDir)
		} else if DoesDirExist(depDir) && DoesDirExist(newDir) { //both exist, better check what's in them
			if err := checkFileNamesAndMigrate(depDir, newDir); err != nil {
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
		depFile := path.Join(depDir, file.Name())
		newFile := path.Join(newDir, file.Name()) //file may not actually exist (yet)

		if fileNamesToCheck[file.Name()] == true { //conflict!
			return fmt.Errorf("identical file name in deprecated dir (%s) and new dir to migrate to (%s)\nplease resolve and re-run command", depFile, newFile)
		} else { //filenames don't match, move file from depDir to newDir
			logger.Printf("File migration NOT succesful:\t%s ====> %s\n", depFile, newFile)

			if err := os.Rename(depFile, newFile); err != nil {
				return err
			}
			logger.Printf("File migration succesful:\t%s ====> %s\n", depFile, newFile)
		}
	}
	return nil
}

func DoesDirExist(dir string) bool {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		//logger.Debugf("%s does not exist\n", dir)
		return false
	} else {
		//logger.Debugf("%s does exist\n", dir)
		return true
	}
}
