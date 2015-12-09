package initialize

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/eris-ltd/eris-cli/util"
	"github.com/eris-ltd/eris-cli/version"

	log "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/common"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/ipfs"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"
)

//TODO hard code in /common
var defChainDir = path.Join(common.ChainsPath, "default")

func dropServiceDefaults(dir, from string) error {
	servDefs := []string{
		"btcd.toml",
		"eth.toml",
		"ipfs.toml",
		"keys.toml",
		"mindy.toml",
		"mint.toml",
		"openbazaar.toml",
		"tinydns.toml",
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
	return nil
}

func dropChainDefaults(dir, from string) error {
	if err := os.MkdirAll(defChainDir, 0777); err != nil {
		return err
	}

	chnDefs := []string{
		"default.toml",
		"config.toml",
		"server_conf.toml",
	}
	if err := drops(chnDefs, "chains", dir, from); err != nil {
		return err
	}
	if err := writeDefaultFile(defChainDir, "genesis.json", DefChainGen); err != nil {
		return fmt.Errorf("Cannot add default genesis.json: %s.\n", err)
	}
	if err := writeDefaultFile(defChainDir, "priv_validator.json", DefChainKeys); err != nil {
		return fmt.Errorf("Cannot add default priv_validator.json: %s.\n", err)
	}
	if err := writeDefaultFile(defChainDir, "genesis.csv", DefChainCSV); err != nil {
		return fmt.Errorf("Cannot add default genesis.csv: %s.\n", err)
	}

	versionDefault := path.Join(dir, "default.toml")

	read, err := ioutil.ReadFile(versionDefault)
	if err != nil {
		return err
	}

	//TODO update eris-chains/default.toml for only one "version" in the file
	withVersion := strings.Replace(string(read), "version", version.VERSION, 2)

	if err := ioutil.WriteFile(versionDefault, []byte(withVersion), 0); err != nil {
		return err
	}

	//move things to where they ought to be
	config := path.Join(dir, "config.toml")
	configDef := path.Join(defChainDir, "config.toml")
	if err := os.Rename(config, configDef); err != nil {
		return err
	}

	server := path.Join(dir, "server_conf.toml")
	serverDef := path.Join(defChainDir, "server_conf.toml")
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

	logger.Println("Pulling default docker images from quay.io")
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
		auth := docker.AuthConfiguration{}

		if err := util.DockerClient.PullImage(opts, auth); err != nil {
			return err
		}
	}
	return nil
}

func drops(files []string, typ, dir, from string) error {
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
			logger.Debugf("Getting %s from:\t%s\n", file, url)
			if err := ipfs.DownloadFromUrlToFile(url, file, dir, buf); err != nil {
				return err
			}
		}
	} else if from == "rawgit" {
		for _, file := range files {
			logger.Debugf("Getting %s from: GitHub\n", file)
			//TODO deduplicate that dum file
			if err := util.GetFromGithub("eris-ltd", repo, "master", file, dir, file, buf); err != nil {
				return err
			}
		}
	}
	return nil
}

//XXX legacy - keep around for good measure :~)
//func dropChainDefaults(dir, from string) error {
//	defChainDir := filepath.Join(common.ChainsPath, "default")
//	return nil
//}

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
