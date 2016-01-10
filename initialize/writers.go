package initialize

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/eris-ltd/eris-cli/util"
	"github.com/eris-ltd/eris-cli/version"

	log "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/common"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/ipfs"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"
)

// XXX all files in this sequence must be added to both
// the respective GH repo & mindy testnet (pinkpenguin.interblock.io:46657/list_name)
func dropServiceDefaults(dir, from string) error {
	servDefs := []string{
		"btcd.toml",
		"compilers.toml",
		"eth.toml",
		"ipfs.toml",
		"keys.toml",
		"logspout.toml",
		"logsrotate.toml",
		"mindy.toml",
		"mint.toml",
		"openbazaar.toml",
		"tinydns.toml",
		"watchtower.toml",
		"do_not_use.toml",
	}

	if err := drops(servDefs, "services", dir, from); err != nil {
		return err
	}
	return nil
}

func dropActionDefaults(dir, from string) error {
	actDefs := []string{
		"chain_info.toml",
		"dns_register.toml",
		"keys_list.toml",
	}
	if err := drops(actDefs, "actions", dir, from); err != nil {
		return err
	}
	if err := writeDefaultFile(common.ActionsPath, "do_not_use.toml", defAct); err != nil {
		return fmt.Errorf("Cannot add default do_not_use: %s.\n", err)
	}
	return nil
}

func dropChainDefaults(dir, from string) error {
	chnDefs := []string{
		"default.toml",
		"config.toml",
		"server_conf.toml",
	}
	if err := drops(chnDefs, "chains", dir, from); err != nil {
		return err
	}

	// common.DefaultChainDir goes to /home/zach/.eris
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
	withVersion := strings.Replace(string(read), "version", version.VERSION, 2)
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
		"quay.io/eris/base",
		"quay.io/eris/keys",
		"quay.io/eris/data",
		"quay.io/eris/ipfs",
		"quay.io/eris/erisdb",
		"quay.io/eris/epm",
	}

	log.Warn("Pulling default docker images from quay.io")
	for _, image := range images {
		var tag string
		if image == "eris/erisdb" || image == "eris/epm" {
			tag = version.VERSION
		} else {
			tag = "latest"
		}
		opts := docker.PullImageOptions{
			Repository:   image,
			Registry:     "quay.io",
			Tag:          tag,
			OutputStream: os.Stdout,
		}
		if os.Getenv("ERIS_PULL_APPROVE") == "true" {
			opts.OutputStream = nil
		}

		auth := docker.AuthConfiguration{}

		if err := util.DockerClient.PullImage(opts, auth); err != nil {
			return err
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
