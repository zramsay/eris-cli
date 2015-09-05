package initialize

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/eris-ltd/eris-cli/config"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/common"
)

func cloneRepo(prompt bool, name, location string) error {
	src := "https://github.com/eris-ltd/" + name

	if _, err := os.Stat(filepath.Join(location, ".git")); !os.IsNotExist(err) {

		// if the .git directory exists within ~/.eris/services (or w/e)
		//   then pull rather than clone.
		logger.Printf("The location is a git repository. Attempting to pull instead.\n")
		if err := pullRepo(location, prompt); err != nil {
			return err
		}

	} else {

		// if the ~/.eris/services (or w/e) directory exists, but it does
		// not have a .git directory (caught above), then clone the repo
		if prompt || askToPull(location) {

			// if users want to clone the repository we clear the directory first to avoid
			//   cannot merge errors on localized changes. This is very opinionated and may
			//   need to change down the road. generally, this should not be a big problem.
			if e2 := common.ClearDir(location); e2 != nil {
				return e2
			}

			logger.Printf("Cloning git repository.\n")
			c := exec.Command("git", "clone", src, location)

			// XXX [csk]: we squelch this output to provide a nicer newb interface... may need to change later
			//c.Stdout = config.GlobalConfig.Writer
			c.Stderr = config.GlobalConfig.ErrorWriter
			if e3 := c.Run(); e3 != nil {
				return e3
			}

		} else {
			logger.Debugf("Authorization not granted. Skipping.\n")
		}
	}
	return nil
}

func pullRepo(location string, alreadyAsked bool) error {
	if alreadyAsked || askToPull(location) {
		prevDir, _ := os.Getwd()

		if err := os.Chdir(location); err != nil {
			return fmt.Errorf("Error:\tCould not move into the directory (%s)\n", location)
		}

		logger.Printf("Pulling origin master.\n")
		c := exec.Command("git", "pull", "origin", "master")

		// XXX [csk]: we squelch this output to provide a nicer newb interface... may need to change later
		//c.Stdout = config.GlobalConfig.Writer

		c.Stderr = config.GlobalConfig.ErrorWriter
		if err := c.Run(); err != nil {
			logger.Printf("err: %v\n", err)
			return err
		}

		if err := os.Chdir(prevDir); err != nil {
			return fmt.Errorf("Error:\tCould not move into the directory (%s)\n", prevDir)
		}
	}

	return nil
}

func askToPull(location string) bool {
	var input string

	logger.Printf("Looks like the %s directory exists.\nWould you like the marmots to pull in any recent changes? (Y/n): ", location)
	fmt.Scanln(&input)

	if input == "Y" || input == "y" || input == "YES" || input == "Yes" || input == "yes" {
		return true
	}
	return false
}

func dropDefaults() error {
	if err := writeDefaultFile(common.ServicesPath, "keys.toml", DefaultKeys); err != nil {
		return fmt.Errorf("Cannot add keys: %s.\n", err)
	}
	if err := writeDefaultFile(common.ServicesPath, "ipfs.toml", DefaultIpfs); err != nil {
		return fmt.Errorf("Cannot add ipfs: %s.\n", err)
	}
	if err := writeDefaultFile(common.ServicesPath, "do_not_use.toml", DefaultIpfs2); err != nil {
		return fmt.Errorf("Cannot add ipfs: %s.\n", err)
	}
	if err := writeDefaultFile(common.ActionsPath, "do_not_use.toml", defAct); err != nil {
		return fmt.Errorf("Cannot add default action: %s.\n", err)
	}
	return nil
}

func dropChainDefaults() error {
	defChainDir := filepath.Join(common.BlockchainsPath, "config", "default")
	if err := writeDefaultFile(common.BlockchainsPath, "default.toml", DefChainService); err != nil {
		return fmt.Errorf("Cannot add default chain definition: %s.\n", err)
	}
	if err := writeDefaultFile(defChainDir, "config.toml", DefChainConfig); err != nil {
		return fmt.Errorf("Cannot add default config.toml: %s.\n", err)
	}
	if err := writeDefaultFile(defChainDir, "genesis.json", DefChainGen); err != nil {
		return fmt.Errorf("Cannot add default genesis.json: %s.\n", err)
	}
	if err := writeDefaultFile(defChainDir, "priv_validator.json", DefChainKeys); err != nil {
		return fmt.Errorf("Cannot add default priv_validator.json: %s.\n", err)
	}
	if err := writeDefaultFile(defChainDir, "server_conf.toml", DefChainServConfig); err != nil {
		return fmt.Errorf("Cannot add default server_conf.toml: %s.\n", err)
	}
	if err := writeDefaultFile(defChainDir, "genesis.csv", DefChainCSV); err != nil {
		return fmt.Errorf("Cannot add default genesis.csv: %s.\n", err)
	}
	return nil
}

func writeDefaultFile(savePath, fileName string, toWrite func() string) error {
	if err := os.MkdirAll(savePath, 0777); err != nil {
		return err
	}
	writer, err := os.Create(filepath.Join(savePath, fileName))
	defer writer.Close()
	if err != nil {
		return err
	}
	writer.Write([]byte(toWrite()))
	return nil
}
