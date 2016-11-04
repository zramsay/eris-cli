package keys

import (
	"bytes"
	"io"
	"io/ioutil"
	"path"
	"path/filepath"
	"strings"

	"github.com/eris-ltd/eris-cli/config"
	"github.com/eris-ltd/eris-cli/data"
	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/log"
	"github.com/eris-ltd/eris-cli/services"
)

func ListKeys(do *definitions.Do) ([]string, error) {
	var result []string
	if do.Host {
		keysPath := filepath.Join(config.KeysPath, "data")
		addrs, err := ioutil.ReadDir(keysPath)
		if err != nil {
			return nil, err
		}
		for _, addr := range addrs {
			result = append(result, addr.Name())
		}
		if !do.Quiet {
			if len(addrs) == 0 {
				log.Warn("No keys found on host")
			} else {
				// First key.
				log.WithField("=>", result[0]).Warn("The keys on your host kind marmot")
				// Subsequent keys.
				if len(result) > 1 {
					for _, addr := range result[1:] {
						log.WithField("=>", addr).Warn()
					}
				}
			}
		}
	}

	if do.Container {
		do.Name = "keys"
		if err := services.EnsureRunning(do); err != nil {
			return nil, err
		}

		keysOut, err := services.ExecHandler(do.Name, []string{"ls", "/home/eris/.eris/keys/data"})
		if err != nil {
			return nil, err
		}
		result = strings.Fields(keysOut.String())
		if !do.Quiet {
			if len(result) == 0 || result[0] == "" {
				log.Warn("No keys found in container")
			} else {
				// First key.
				log.WithField("=>", result[0]).Warn("The keys in your container kind marmot")
				// Subsequent keys.
				if len(result) > 1 {
					for _, addr := range result[1:] {
						log.WithField("=>", addr).Warn()
					}
				}
			}
		}
	}
	return result, nil
}

func GenerateKey(do *definitions.Do) error {
	do.Name = "keys"

	if err := services.EnsureRunning(do); err != nil {
		return err
	}
	// TODO implement
	// if do.Password {}

	buf, err := services.ExecHandler(do.Name, []string{"eris-keys", "gen", "--no-pass"})
	if err != nil {
		return err
	}

	if do.Save {
		addr := new(bytes.Buffer)
		addr.ReadFrom(buf)

		doExport := definitions.NowDo()
		doExport.Address = strings.TrimSpace(addr.String())

		log.WithField("=>", doExport.Address).Warn("Saving key to host")
		if err := ExportKey(doExport); err != nil {
			return err
		}
	}

	io.Copy(config.Global.Writer, buf)

	return nil
}

func ExportKey(do *definitions.Do) error {
	do.Name = "keys"
	if err := services.EnsureRunning(do); err != nil {
		return err
	}

	if do.All && do.Address == "" {
		do.Destination = config.KeysPath
		do.Source = path.Join(config.KeysContainerPath)
	} else {
		do.Destination = config.KeysDataPath
		do.Source = path.Join(config.KeysContainerPath, do.Address)
	}
	return data.ExportData(do)
}

func ImportKey(do *definitions.Do) error {
	do.Name = "keys"
	if err := services.EnsureRunning(do); err != nil {
		return err
	}

	if do.All && do.Address == "" {
		doLs := definitions.NowDo()
		doLs.Container = false
		doLs.Host = true
		doLs.Quiet = true
		result, err := ListKeys(doLs)
		if err != nil {
			return err
		}

		for _, addr := range result {
			do.Source = filepath.Join(config.KeysDataPath, addr)
			do.Destination = path.Join(config.KeysContainerPath, addr)
			if err := data.ImportData(do); err != nil {
				return err
			}
		}
	} else {
		do.Source = filepath.Join(config.KeysDataPath, do.Address)
		do.Destination = path.Join(config.KeysContainerPath, do.Address)
		if err := data.ImportData(do); err != nil {
			return err
		}
	}

	return nil
}
