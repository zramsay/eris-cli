package keys

import (
	"bytes"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/eris-ltd/eris-cli/config"
	"github.com/eris-ltd/eris-cli/definitions"
	srv "github.com/eris-ltd/eris-cli/services"
	"github.com/eris-ltd/eris-cli/util"

	log "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	. "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/common"
)

func ListKeys(do *definitions.Do) error {
	if do.Host {
		keysPath := filepath.Join(KeysPath, "data")
		addrs, err := ioutil.ReadDir(keysPath)
		if err != nil {
			return err
		}
		hostAddrs := make([]string, len(addrs))
		for i, addr := range addrs {
			hostAddrs[i] = addr.Name()
		}
		log.WithField("=>", hostAddrs[0]).Warn("The keys on your host kind marmot:")
		hostAddrs = append(hostAddrs[:0], hostAddrs[1:]...)
		for _, addr := range hostAddrs {
			log.WithField("=>", addr).Warn()
		}
	}

	if do.Container {
		keysOut := new(bytes.Buffer)
		config.GlobalConfig.Writer = keysOut
		do.Name = "keys"
		do.Operations.ContainerNumber = 1
		if err := srv.EnsureRunning(do); err != nil {
			return err
		}

		do.Operations.Interactive = false
		do.Operations.Args = []string{"ls", "/home/eris/.eris/keys/data"}

		if err := srv.ExecService(do); err != nil {
			return err
		}
		keysOutString := strings.Split(util.TrimString(string(keysOut.Bytes())), "\n")
		log.WithField("=>", keysOutString[0]).Warn("The keys in your container kind marmot:")
		keysOutString = append(keysOutString[:0], keysOutString[1:]...)
		for _, addr := range keysOutString {
			log.WithField("=>", addr).Warn()
		}

	}
	return nil
}
