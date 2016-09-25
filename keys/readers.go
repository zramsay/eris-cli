package keys

import (
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/eris-ltd/eris-cli/definitions"
	srv "github.com/eris-ltd/eris-cli/services"

	. "github.com/eris-ltd/common/go/common"
	log "github.com/eris-ltd/eris-logger"
)

func ListKeys(do *definitions.Do) ([]string, error) {
	var result []string
	if do.Host {
		keysPath := filepath.Join(KeysPath, "data")
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
		if err := srv.EnsureRunning(do); err != nil {
			return nil, err
		}

		keysOut, err := srv.ExecHandler(do.Name, []string{"ls", "/home/eris/.eris/keys/data"})
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
