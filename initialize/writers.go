package initialize

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/common"
)

func pullRepo(name, location string, verbose bool) error {
	src := "https://github.com/eris-ltd/" + name
	c := exec.Command("git", "clone", src, location)
	if verbose {
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
	}
	if err := c.Run(); err != nil {
		return err
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
	if err := writeDefaultFile(defChainDir, "config.toml", defChainConfig); err != nil {
		return fmt.Errorf("Cannot add default config.toml: %s.\n", err)
	}
	if err := writeDefaultFile(defChainDir, "genesis.json", defChainGen); err != nil {
		return fmt.Errorf("Cannot add default genesis.json: %s.\n", err)
	}
	if err := writeDefaultFile(defChainDir, "priv_validator.json", defChainKeys); err != nil {
		return fmt.Errorf("Cannot add default priv_validator.json: %s.\n", err)
	}
	if err := writeDefaultFile(defChainDir, "server_conf.toml", defChainServConfig); err != nil {
		return fmt.Errorf("Cannot add default server_conf.toml: %s.\n", err)
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
	def := toWrite()
	writer.Write([]byte(def))
	return nil
}
