package keys

import (
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/eris-ltd/eris-cli/definitions"
	srv "github.com/eris-ltd/eris-cli/services"
	"github.com/eris-ltd/eris-cli/util"
	. "github.com/eris-ltd/eris-cli/errors"

	log "github.com/eris-ltd/eris-logger"
	. "github.com/eris-ltd/common/go/common"
)

func ListKeys(do *definitions.Do) error {
	if do.Host {
		keysPath := filepath.Join(KeysPath, "data")
		addrs, err := ioutil.ReadDir(keysPath)
		if err != nil {
			return &ErisError{ErrGo, err, "check your keys path"}
		}
		if !do.Quiet { // [zr] ... ?
			if len(addrs) == 0 {
				log.Warn("No keys found on host")
			} else {
				hostAddrs := make([]string, len(addrs))
				for i, addr := range addrs {
					hostAddrs[i] = addr.Name()
				}
				do.Result = strings.Join(hostAddrs, ",")
				log.WithField("=>", hostAddrs[0]).Warn("The keys on your host kind marmot")
				hostAddrs = append(hostAddrs[:0], hostAddrs[1:]...)
				for _, addr := range hostAddrs {
					log.WithField("=>", addr).Warn()
				}
			}
		}
	}

	if do.Container {
		do.Name = "keys"
		if err := srv.EnsureRunning(do); err != nil {
			return &ErisError{ErrEris, err, "[eris services start keys]"}
		}

		keysOut, err := srv.ExecHandler(do.Name, []string{"ls", "/home/eris/.eris/keys/data"})
		if err != nil {
			return &ErisError{ErrDocker, err, "check that..."}
		}
		keysOutString := strings.Split(util.TrimString(keysOut.String()), "\n")
		do.Result = strings.Join(keysOutString, ",")
		if !do.Quiet {
			if len(keysOutString) == 0 || keysOutString[0] == "" {
				log.Warn("No keys found in container")
			} else {
				log.WithField("=>", keysOutString[0]).Warn("The keys in your container kind marmot")
				keysOutString = append(keysOutString[:0], keysOutString[1:]...)
				for _, addr := range keysOutString {
					log.WithField("=>", addr).Warn()
				}
			}
		}
	}
	return nil
}
