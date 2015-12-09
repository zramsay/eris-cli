package ipfs

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
	"strings"
	"time"
)

func GetFromIPFS(hash, fileName, dirName string, w io.Writer) error {
	url := IPFSBaseGatewayUrl("") + hash
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
		fmt.Printf("Downloading %s to %s\n", url, endPath)
		checkDir, err := os.Stat(dirName)
		if err != nil {
			return err
		}
		if !checkDir.IsDir() {
			w.Write([]byte("Directory does not exist, creating it"))
			err1 := os.MkdirAll(dirName, 0700)
			if err1 != nil {
				return fmt.Errorf("error making directory, check your permissions %v\n", err1)
			}
		}
		//	return fmt.Errorf("path specified is not a directory, please enter a directory")
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
	if string(checkBody) == "Path Resolve error: context deadline exceeded" {
		//this won't work unless we `eris files cache --csv (which will be especially needed to deal with this error)
		return fmt.Errorf("A timeout occured while trying to reach IPFS. Run `eris files cache [hash], wait 5-10 seconds, then run `eris files [cmd] [hash]`")
	}

	return nil
}

// --------------------------------------------------------------
// Helper functions

var timeout = time.Duration(10 * time.Second)

func dialTimeout(network, addr string) (net.Conn, error) {
	return net.DialTimeout(network, addr, timeout)
}
