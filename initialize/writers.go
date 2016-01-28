package initialize

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/eris-ltd/eris-cli/perform"
	"github.com/eris-ltd/eris-cli/util"
	ver "github.com/eris-ltd/eris-cli/version"

	log "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/common"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/ipfs"
	//"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"
)

// XXX all files in this sequence must be added to both
// the respective GH repo & mindy testnet (pinkpenguin.interblock.io:46657/list_names)
func dropServiceDefaults(dir, from string) error {
	if err := drops(ver.SERVICE_DEFINITIONS, "services", dir, from); err != nil {
		return err
	}
	return nil
}

func dropActionDefaults(dir, from string) error {
	if err := drops(ver.ACTION_DEFINITIONS, "actions", dir, from); err != nil {
		return err
	}
	if err := writeDefaultFile(common.ActionsPath, "do_not_use.toml", defAct); err != nil {
		return fmt.Errorf("Cannot add default do_not_use: %s.\n", err)
	}
	return nil
}

func dropChainDefaults(dir, from string) error {
	if err := drops(ver.CHAIN_DEFINITIONS, "chains", dir, from); err != nil {
		return err
	}

	// common.DefaultChainDir goes to $HOME/.eris
	// rather than /tmp/eris/.eris
	// XXX something wonky with ResolveErisRoot()?
	chnDir := filepath.Join(dir, "default")
	if err := writeDefaultFile(chnDir, "genesis.json", DefChainGen); err != nil {
		return fmt.Errorf("Cannot add default genesis.json: %s.\n", err)
	}
	if err := writeDefaultFile(chnDir, "priv_validator.json", DefChainKeys); err != nil {
		return fmt.Errorf("Cannot add default priv_validator.json: %s.\n", err)
	}
	if err := writeDefaultFile(chnDir, "genesis.csv", DefChainCSV); err != nil {
		return fmt.Errorf("Cannot add default genesis.csv: %s.\n", err)
	}

	//insert version into default chain service definition
	versionDefault := filepath.Join(dir, "default.toml")
	read, err := ioutil.ReadFile(versionDefault)
	if err != nil {
		return err
	}
	withVersion := strings.Replace(string(read), "version", ver.VERSION, 2)
	if err := ioutil.WriteFile(versionDefault, []byte(withVersion), 0); err != nil {
		return err
	}

	//move things to where they ought to be
	config := filepath.Join(dir, "config.toml")
	configDef := filepath.Join(chnDir, "config.toml")
	if err := os.Rename(config, configDef); err != nil {
		return err
	}

	server := filepath.Join(dir, "server_conf.toml")
	serverDef := filepath.Join(chnDir, "server_conf.toml")
	if err := os.Rename(server, serverDef); err != nil {
		return err
	}
	return nil
}

func pullDefaultImages() error {
	images := []string{
		path.Join(ver.QUAY, ver.ERIS_IMG_BASE),
		path.Join(ver.QUAY, ver.ERIS_IMG_DATA),
		path.Join(ver.QUAY, ver.ERIS_IMG_KEYS),
		path.Join(ver.QUAY, ver.ERIS_IMG_IPFS),
		path.Join(ver.QUAY, ver.ERIS_IMG_DB),
		path.Join(ver.QUAY, ver.ERIS_IMG_PM),
	}

	log.Warn("Pulling default docker images from quay.io")
	for _, image := range images {
		//XXX can'tuse b/c import cycle :(
		if err := perform.PullImage(image, nil); err != nil {

		}
	}
	return nil
}

func drops(files []string, typ, dir, from string) error {
	//to get from rawgit
	var repo string
	if typ == "services" {
		repo = "eris-services"
	} else if typ == "actions" {
		repo = "eris-actions"
	} else if typ == "chains" {
		repo = "eris-chains"
	}

	if !util.DoesDirExist(dir) {
		if err := os.MkdirAll(dir, 0777); err != nil {
			return err
		}
	}

	buf := new(bytes.Buffer)
	if from == "toadserver" {
		for _, file := range files {
			url := fmt.Sprintf("%s:11113/getfile/%s", ipfs.SexyUrl(), file)
			log.WithField(file, url).Debug("Getting file from url")
			log.WithField(file, dir).Debug("Dropping file to")
			if err := ipfs.DownloadFromUrlToFile(url, file, dir, buf); err != nil {
				return err
			}
		}
	} else if from == "rawgit" {
		for _, file := range files {
			log.WithField(file, dir).Debug("Getting file from GitHub, dropping into:")
			if err := util.GetFromGithub("eris-ltd", repo, "master", file, dir, file, buf); err != nil {
				return err
			}
		}
	}
	return nil
}

//TODO eventually eliminate this
func writeDefaultFile(savePath, fileName string, toWrite func() string) error {
	if err := os.MkdirAll(savePath, 0777); err != nil {
		return err
	}
	pth := filepath.Join(savePath, fileName)
	writer, err := os.Create(pth)
	defer writer.Close()
	if err != nil {
		return err
	}
	writer.Write([]byte(toWrite()))
	return nil
}
