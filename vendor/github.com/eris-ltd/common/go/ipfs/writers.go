package ipfs

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	log "github.com/eris-ltd/eris-logger"
)

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
			return "", fmt.Errorf("The file has already been pinned recusively (probably from ipfs add or eris files put). It is only possible to cache a file you don't already have. see issue #133 for more information")
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
