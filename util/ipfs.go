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
	"strings"
	"time"

	"github.com/monax/cli/log"
)

var (
	IpfsHost string = "http://0.0.0.0"
	IpfsPort string = "8080"

	timeout = time.Duration(10 * time.Second)
)

func IPFSBaseGatewayUrl(gateway, port string) string {
	if port == "" {
		port = IpfsPort
	}
	if gateway == "monax" {
		return fmt.Sprintf("%s:%s%s", SexyUrl(), port, "/ipfs/")
	} else if gateway != "" {
		return fmt.Sprintf("%s:%s%s", gateway, port, "/ipfs/")
	} else {
		return fmt.Sprintf("%s:%s%s", IPFSUrl(), port, "/ipfs/")
	}
}

func IPFSBaseAPIUrl() string {
	return fmt.Sprintf("%s%s", IPFSUrl(), ":5001/api/v0/")
}

func SexyUrl() string {
	//TODO load balancer (when one isn't enough)
	return "http://ipfs.monax.io"
}

func IPFSUrl() string {
	var host string
	if os.Getenv("MONAX_CLI_CONTAINER") == "true" {
		host = "http://ipfs"
	} else {
		if os.Getenv("MONAX_IPFS_HOST") != "" {
			host = os.Getenv("MONAX_IPFS_HOST")
		} else {
			host = IpfsHost
		}
	}
	return host
}

func GetFromIPFS(hash, fileName, dirName, port string) error {
	url := IPFSBaseGatewayUrl("", port) + hash // [csk] why isn't this using a gateway argument like the Put?
	log.WithFields(log.Fields{
		"file": fileName,
		"hash": hash,
	}).Warn("Getting file from IPFS")
	return DownloadFromUrlToFile(url, fileName, dirName)
}

func CatFromIPFS(fileHash string) (string, error) {
	url := IPFSBaseAPIUrl() + "cat?arg=" + fileHash
	log.WithFields(log.Fields{
		"hash": fileHash,
	}).Warn("Catting file from IPFS")
	body, err := PostAPICall(url, fileHash)

	if err != nil {
		return "", err
	}

	return string(body), nil
}

func ListFromIPFS(objectHash string) (string, error) {
	url := IPFSBaseAPIUrl() + "ls?arg=" + objectHash
	log.WithFields(log.Fields{
		"hash": objectHash,
	}).Warn("Listing file from IPFS")
	body, err := PostAPICall(url, objectHash)
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

func ListPinnedFromIPFS() (string, error) {
	url := IPFSBaseAPIUrl() + "pin/ls"
	log.Warn("Listing files pinned locally")
	body, err := PostAPICall(url, "")
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

func DownloadFromUrlToFile(url, fileName, dirName string) error {
	tokens := strings.Split(url, "/")
	if fileName == "" {
		fileName = tokens[len(tokens)-1]
	}

	//use absolute paths?
	endPath := path.Join(dirName, fileName)
	if dirName != "" {
		log.WithFields(log.Fields{
			"from": url,
			"to":   endPath,
		}).Info("Downloading")
		checkDir, err := os.Stat(dirName)
		if err != nil {
			log.Warn("Directory does not exist, creating it")
			err1 := os.MkdirAll(dirName, 0700)
			if err1 != nil {
				return fmt.Errorf("error making directory, check your permissions %v\n", err1)
			}
		}
		if !checkDir.IsDir() {
			return fmt.Errorf("path specified is not a directory, please enter a directory")
		}
	} else {
		log.WithFields(log.Fields{
			"from": url,
			"to":   fileName,
		}).Warn("Downloading")
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
		//this won't work unless we `monax files cache --csv (which will be especially needed to deal with this error)
		return fmt.Errorf("A timeout occured while trying to reach IPFS. Run `monax files cache [hash], wait 5-10 seconds, then run `monax files [cmd] [hash]`")
	}

	return nil
}

func SetLogLevel(level log.Level) {
	log.SetLevel(level)
}

func SendToIPFS(fileName, gateway, port string) (string, error) {
	url := IPFSBaseGatewayUrl(gateway, port)
	log.WithField("file", fileName).Warn("Posting file to IPFS")
	head, err := UploadFromFileToUrl(url, fileName)
	if err != nil {
		return "", err
	}
	hash, ok := head["Ipfs-Hash"]
	if !ok || hash[0] == "" {
		return "", fmt.Errorf("No hash returned")
	}
	return hash[0], nil
}

func PinToIPFS(fileHash string) (string, error) {
	url := IPFSBaseAPIUrl() + "pin/add?arg=" + fileHash
	log.WithField("hash", fileHash).Warn("Pinning file to IPFS")
	body, err := PostAPICall(url, fileHash)
	if err != nil {
		return "", err
	}
	log.WithFields(log.Fields{
		"hash": fileHash,
		"url":  url,
	}).Warn("Caching")
	var p struct {
		Pinned []string
	}

	if err = json.Unmarshal(body, &p); err != nil || len(p.Pinned) == 0 {
		//XXX hacky
		if fmt.Sprintf("%v", err) == "invalid character 'p' looking for beginning of value" {
			return "", fmt.Errorf("The file has already been pinned recusively (probably from ipfs add or monax files put). It is only possible to cache a file you don't already have. see issue #133 for more information")
		}
		return "", fmt.Errorf("unexpected error unmarshalling json: %v", err)
	}

	return p.Pinned[0], nil
}

func RemovePinnedFromIPFS(fileHash string) (string, error) {
	url := IPFSBaseAPIUrl() + "pin/rm?arg=" + fileHash

	log.WithField("hash", fileHash).Warn("Removing pinned file to IPFS")
	body, err := PostAPICall(url, fileHash)
	if err != nil {
		return "", err
	}
	log.WithFields(log.Fields{
		"hash": fileHash,
		"url":  url,
	}).Warn("Deleting from cache")
	var p struct {
		Pinned []string
	}

	if err = json.Unmarshal(body, &p); err != nil {
		return "", fmt.Errorf("unexpected error unmarshalling json: %v", err)
	}

	return p.Pinned[0], nil
}

func UploadFromFileToUrl(url, fileName string) (http.Header, error) {
	if url == "" {
		return http.Header{}, fmt.Errorf("To upload from a file to a url, I need a URL.")
	}
	log.WithFields(log.Fields{
		"file": fileName,
		"url":  url,
	}).Warn("Uploading")

	input, err := os.Open(fileName)
	if err != nil {
		return http.Header{}, err
	}
	defer input.Close()

	request, err := http.NewRequest("POST", url, input)
	if err != nil {
		return http.Header{}, err
	}
	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return http.Header{}, err
	}
	defer response.Body.Close()

	return response.Header, nil
}

//returns []byte to let each command make its own struct for the response
//but handles the errs in here
func PostAPICall(url, fileHash string) ([]byte, error) {
	request, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return []byte(""), err
	}
	request.Close = true //for successive api calls
	request.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return []byte(""), err
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return []byte(""), err
	}

	var errs struct {
		Message string
		Code    int
	}
	if response.StatusCode >= http.StatusBadRequest {
		//TODO better err handling; this is a (very) slimed version of how IPFS does it.
		if err = json.Unmarshal(body, &errs); err != nil {
			return []byte(""), fmt.Errorf("error json unmarshaling body (bad request): %v", err)
		}
		return []byte(errs.Message), nil

		if response.StatusCode == http.StatusNotFound {
			if err = json.Unmarshal(body, &errs); err != nil {
				return []byte(""), fmt.Errorf("error json unmarshaling body (status not found): %v", err)
			}
			return []byte(errs.Message), nil
		}
	}
	//XXX hacky: would need to fix ipfs error msgs
	if string(body) == "Path Resolve error: context deadline exceeded" && string(body) == "context deadline exceeded" {
		return []byte(""), fmt.Errorf("A timeout occured while trying to reach IPFS. Run `[monax files cache HASH]`, wait 5-10 seconds, then run `[monax files COMMAND HASH]`")
	}
	return body, nil
}

func dialTimeout(network, addr string) (net.Conn, error) {
	return net.DialTimeout(network, addr, timeout)
}
