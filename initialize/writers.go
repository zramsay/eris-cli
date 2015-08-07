package initialize

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common"
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
	if err := keysDef(); err != nil {
		return fmt.Errorf("Cannot add keys: %s.\n", err)
	}
	if err := ipfsDef(); err != nil {
		return fmt.Errorf("Cannot add ipfs: %s.\n", err)
	}
	// if err := edbDef(); err != nil {
	// 	return fmt.Errorf("Cannot add erisdb: %s.\n", err)
	// }
	if err := genDef(); err != nil {
		return fmt.Errorf("Cannot add default genesis: %s.\n", err)
	}
	if err := actDef(); err != nil {
		return fmt.Errorf("Cannot add default action: %s.\n", err)
	}
	return nil
}

func keysDef() error {
	if err := os.MkdirAll(common.ServicesPath, 0777); err != nil {
		return err
	}
	writer, err := os.Create(filepath.Join(common.ServicesPath, "keys.toml"))
	defer writer.Close()
	if err != nil {
		return err
	}
	keysD := defKeys()
	writer.Write([]byte(keysD))
	return nil
}

func ipfsDef() error {
	if err := os.MkdirAll(common.ServicesPath, 0777); err != nil {
		return err
	}
	writer, err := os.Create(filepath.Join(common.ServicesPath, "ipfs.toml"))
	defer writer.Close()
	if err != nil {
		return err
	}
	ipfsD := defIpfs()
	writer.Write([]byte(ipfsD))
	return nil
}

func edbDef() error {
	if err := os.MkdirAll(common.ServicesPath, 0777); err != nil {
		return err
	}
	writer, err := os.Create(filepath.Join(common.ServicesPath, "erisdb.toml"))
	defer writer.Close()
	if err != nil {
		return err
	}
	edbD := defEdb()
	writer.Write([]byte(edbD))
	return nil
}

func genDef() error {
	genPath := filepath.Join(common.BlockchainsPath, "genesis")
	if err := os.MkdirAll(genPath, 0777); err != nil {
		return err
	}
	writer, err := os.Create(filepath.Join(genPath, "default.json"))
	defer writer.Close()
	if err != nil {
		return err
	}
	gen := defGen()
	writer.Write([]byte(gen))
	return nil
}

func actDef() error {
	if err := os.MkdirAll(common.ActionsPath, 0777); err != nil {
		return err
	}
	writer, err := os.Create(filepath.Join(common.ActionsPath, "do_not_use.toml"))
	defer writer.Close()
	if err != nil {
		return err
	}
	act := defAct()
	writer.Write([]byte(act))
	return nil
}
