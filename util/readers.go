package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	. "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/common"
)

func GetFromGithub(org, repo, branch, path, fileName string, w io.Writer) error {
	url := "https://rawgit.com/" + strings.Join([]string{org, repo, branch, path}, "/")
	w.Write([]byte("Will download from url -> " + url))
	return DownloadFromUrlToFile(url, fileName, "", w)
}

func GetFromIPFS(hash, fileName, dirName string, w io.Writer) error {
	url := IPFSBaseGatewayUrl(false) + hash
	w.Write([]byte("GETing file from IPFS. Hash =>\t" + hash + ":" + fileName + "\n"))
	return DownloadFromUrlToFile(url, fileName, dirName, w)
}

func CatFromIPFS(fileHash string, w io.Writer) (string, error) {
	url := IPFSBaseAPIUrl() + "cat?arg=" + fileHash
	w.Write([]byte("CATing file from IPFS. Hash =>\t" + fileHash + "\n"))
	body, err := PostAPICall(url, fileHash, w)

	if err != nil {
		return "", err
	}

	return string(body), nil
}

func ListFromIPFS(objectHash string, w io.Writer) (string, error) {
	url := IPFSBaseAPIUrl() + "ls?arg=" + objectHash
	w.Write([]byte("LISTing file from IPFS. objectHash =>\t" + objectHash + "\n"))
	body, err := PostAPICall(url, objectHash, w)
	r := bytes.NewReader(body)

	type LsLink struct {
		Name, Hash string
		Size       uint64
	}
	type LsObject struct {
		Hash  string
		Links []LsLink
	}

	dec := json.NewDecoder(r)
	out := struct{ Objects []LsObject }{}
	err = dec.Decode(&out)
	if err != nil {
		return "", err
	}
	contents := out.Objects[0].Links

	res := make([]string, len(contents))
	for i, c := range contents {
		res[i] = c.Hash + " " + c.Name
	}
	result := strings.Join(res, "\n")
	return result, nil
}

func ListPinnedFromIPFS(w io.Writer) (string, error) {
	url := IPFSBaseAPIUrl() + "pin/ls"
	w.Write([]byte("LISTing files pinned locally.\n"))
	body, err := PostAPICall(url, "", w)
	r := bytes.NewReader(body)

	type RefKeyObject struct {
		Type  string
		Count int
	}

	type RefKeyList struct {
		Keys map[string]RefKeyObject
	}

	var out RefKeyList
	dec := json.NewDecoder(r)
	err = dec.Decode(&out)
	if err != nil {
		return "", err
	}
	contents := out.Keys

	res := make([]string, len(contents))
	i := 0
	for c := range contents {
		res[i] = c
		i += 1
	}
	result := strings.Join(res, "\n")
	return result, nil
}

func DownloadFromUrlToFile(url, fileName, dirName string, w io.Writer) error {
	tokens := strings.Split(url, "/")
	if fileName == "" {
		fileName = tokens[len(tokens)-1]
	}

	//use absolute paths?
	endPath := path.Join(dirName, fileName)
	if dirName != "" {
		w.Write([]byte("Downloading " + url + " to " + endPath + "\n"))
		checkDir, err := os.Stat(dirName)
		if err != nil {
			w.Write([]byte("Directory does not exist, creating it"))
			err1 := os.MkdirAll(dirName, 0700)
			if err1 != nil {
				return fmt.Errorf("error making directory, check your permissions %v\n", err1)
			}
		}
		if !checkDir.IsDir() {
			return fmt.Errorf("path specified is not a directory, please enter a directory")
		}
	} else {
		//dirNAme = getwd
		w.Write([]byte("Downloading " + url + " to " + fileName + "\n"))
	}

	var outputInDir *os.File
	var outputFile *os.File
	var err error
	if dirName != "" {
		outputInDir, err = os.Create(endPath)
		if err != nil {
			return err
		}
		defer outputInDir.Close()
	} else {
		outputFile, err = os.Create(fileName)
		if err != nil {
			return err
		}
		defer outputFile.Close()
	}

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

	var checkBody []byte
	if dirName != "" {
		_, err = io.Copy(outputInDir, response.Body)
		if err != nil {
			return err
		}
		checkBody, err = ioutil.ReadFile(endPath)
		if err != nil {
			return err
		}
	} else {
		_, err = io.Copy(outputFile, response.Body)
		if err != nil {
			return err
		}
		checkBody, err = ioutil.ReadFile(fileName)
		if err != nil {
			return err
		}
	}

	//deal with ipfs' error ungracefully. maybe we want to maintain our own fork?
	//or could run `cache` under the hood, so user doesn't even see error (although we probably shouldn't pin by default)
	//IF this is the only current fix for this error, then we should have `eris files cacheD --rm [hash]` or something
	if string(checkBody) == "Path Resolve error: context deadline exceeded" {
		//this won't work unless we `eris files cache --csv (which will be especially needed to deal with this error)
		return fmt.Errorf("A timeout occured while trying to reach IPFS. Run `eris files cache [hash], wait 5-10 seconds, then run `eris files [cmd] [hash]`")
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

//XXX url funcs can take flags for which host to go to.
func IPFSBaseGatewayUrl(bootstrap bool) string {
	if bootstrap {
		return sexyUrl() + ":8080/ipfs/"
	} else {
		return IPFSUrl() + ":8080/ipfs/"
	}
}

func IPFSBaseAPIUrl() string {
	return IPFSUrl() + ":5001/api/v0/"
}

func sexyUrl() string {
	//bootstrap was down
	//TODO fix before merge; DNS + load balancer
	return "http://147.75.194.73"
}

func IPFSUrl() string {
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
	return host
}

var timeout = time.Duration(10 * time.Second)

func dialTimeout(network, addr string) (net.Conn, error) {
	return net.DialTimeout(network, addr, timeout)
}
