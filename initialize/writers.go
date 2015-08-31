package initialize

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/eris-ltd/eris-cli/config"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/common"
)

func cloneRepo(name, location string) error {
	src := "https://github.com/eris-ltd/" + name
	var c *exec.Cmd

	// if the .git directory exists within ~/.eris/services (or w/e)
	//   then pull rather than clone.
	if _, err := os.Stat(filepath.Join(location, ".git")); !os.IsNotExist(err) {
		logger.Debugf("The location is a git repository. Attempting to pull instead.\n")
		if err := pullRepo(location); err != nil {
			return err
		} else {
			return nil
		}
	}

	// if the ~/.eris/services (or w/e) directory exists, but it does
	//   not have a .git directory (caught above), then init the dir
	//   and pull the repo.
	if _, err := os.Stat(location); !os.IsNotExist(err) {
		logger.Debugf("The location exists but is not a git repository.\nInit-ing git repository.\n")
		c = exec.Command("git", "init", location)
		c.Stdout = config.GlobalConfig.Writer
		c.Stderr = config.GlobalConfig.ErrorWriter
		if e2 := c.Run(); e2 != nil {
			return e2
		}

		logger.Debugf("Adding the proper git remote.\n")
		c = exec.Command("git", "remote", "add", "origin", src)
		if e3 := c.Run(); e3 != nil {
			return e3
		}

		logger.Debugf("Pulling the repository.\n")
		if err := pullRepo(location); err != nil {
			return err
		} else {
			return nil
		}

		// if no ~/.eris/services (or w/e) directory exists, then it will
		//   simply clone in the directory.
	} else {
		c = exec.Command("git", "clone", src, location)
		c.Stdout = config.GlobalConfig.Writer
		c.Stderr = config.GlobalConfig.ErrorWriter
		if err := c.Run(); err != nil {
			return err
		}
	}

	return nil
}

func pullRepo(location string) error {
	var input string
	logger.Printf("Looks like the %s directory exists.\nWould you like the marmots to pull in any recent changes? (Y/n): ", location)
	fmt.Scanln(&input)

	if input == "Y" || input == "y" || input == "YES" || input == "Yes" || input == "yes" {
		prevDir, _ := os.Getwd()
		if err := os.Chdir(location); err != nil {
			return fmt.Errorf("Error:\tCould not move into the directory (%s)\n", location)
		}
		c := exec.Command("git", "pull", "origin", "master")
		c.Stdout = config.GlobalConfig.Writer
		c.Stderr = config.GlobalConfig.ErrorWriter
		if err := c.Run(); err != nil {
			return err
		}
		if err := os.Chdir(prevDir); err != nil {
			return fmt.Errorf("Error:\tCould not move into the directory (%s)\n", location)
		}
	}
	return nil
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
