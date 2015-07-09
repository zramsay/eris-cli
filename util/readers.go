package util

import (
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	. "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common"
)

func GetFromGithub(org, repo, branch, path, fileName string, w io.Writer) error {
	url := "https://rawgit.com/" + strings.Join([]string{org, repo, branch, path}, "/")
	w.Write([]byte("Will download from url -> " + url))
	return DownloadFromUrlToFile(url, fileName, w)
}

func GetFromIPFS(hash, fileName string, w io.Writer) error {
	url := IPFSBaseUrl() + hash
	w.Write([]byte("GETing file from IPFS. Hash =>\t" + hash + ":" + fileName + "\n"))
	return DownloadFromUrlToFile(url, fileName, w)
}

func DownloadFromUrlToFile(url, fileName string, w io.Writer) error {
	tokens := strings.Split(url, "/")
	if fileName == "" {
		fileName = tokens[len(tokens)-1]
	}
	w.Write([]byte("Downloading " + url + " to " + fileName + "\n"))

	output, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer output.Close()

	// adding manual timeouts as IPFS hangs for a while
	transport := http.Transport{
		Dial: dialTimeout,
	}
	client := http.Client{
		Transport: &transport,
	}
	response, err := client.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	_, err = io.Copy(output, response.Body)
	if err != nil {
		return err
	}

	return nil
}

// note this function fails silently.
func GetGlobalLevelConfigFilesByType(typ string, withExt bool) []string {
	var path string
	switch typ {
	case "services":
		path = ServicesPath
	case "chains":
		path = BlockchainsPath
	case "actions":
		path = ActionsPath
	case "files":
		path = FilesPath
	}

	files := []string{}
	fileTypes := []string{}

	// TODO [csk]: DRY up how we deal with file extensions
	for _, t := range []string{"*.json", "*.yaml", "*.toml"} {
		fileTypes = append(fileTypes, filepath.Join(path, t))
	}

	for _, t := range fileTypes {
		s, _ := filepath.Glob(t)
		for _, s1 := range s {
			if !withExt {
				s1 = strings.Split(filepath.Base(s1), ".")[0]
			}
			files = append(files, s1)
		}
	}
	return files
}

func GetFileByNameAndType(typ, name string) string {
	logger.Debugf("Looking for file =>\t\t%s:%s\n", typ, name)
	files := GetGlobalLevelConfigFilesByType(typ, true)

	for _, file := range files {
		file = strings.Split(filepath.Base(file), ".")[0] // quick and dirty file root
		if file == name {
			logger.Debugf("This file found =>\t\t%s\n", file)
			return file
		}
		logger.Debugf("Group file found =>\t\t%s\n", file)
	}

	return ""
}

// --------------------------------------------------------------
// Helper functions

func IPFSBaseUrl() string {
	var host string
	if os.Getenv("ERIS_CLI_CONTAINER") == "true" {
		host = "http://ipfs"
	} else {
		if os.Getenv("ERIS_IPFS_HOST") != "" {
			host = os.Getenv("ERIS_IPFS_HOST")
		} else {
			host = GetConfigValue("IpfsHost")
		}
	}
	return host + ":8080/ipfs/"
}

var timeout = time.Duration(10 * time.Second)

func dialTimeout(network, addr string) (net.Conn, error) {
	return net.DialTimeout(network, addr, timeout)
}
