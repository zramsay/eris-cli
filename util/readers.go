package util

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func GetFromGithub(org, repo, branch, path, fileName string, w io.Writer) error {
	url := "https://rawgit.com/" + strings.Join([]string{org, repo, branch, path}, "/")
	w.Write([]byte("Will download from url -> " + url))
	return DownloadFromUrlToFile(url, fileName, w)
}

func GetFromIPFS(hash, fileName string, w io.Writer) error {
	url := IPFSBaseUrl() + hash
	w.Write([]byte("GETing file from IPFS. Hash -> " + hash))
	return DownloadFromUrlToFile(url, fileName, w)
}

func SendToIPFS(fileName string, w io.Writer) error {
	url := IPFSBaseUrl()
	w.Write([]byte("POSTing file to IPFS. File -> " + fileName))
	return UploadFromFileToUrl(url, fileName, w)
}

func DownloadFromUrlToFile(url, fileName string, w io.Writer) error {
	tokens := strings.Split(url, "/")
	if fileName == "" {
		fileName = tokens[len(tokens)-1]
	}
	w.Write([]byte("Downloading " + url + " to " + fileName))

	output, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer output.Close()

	response, err := http.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	n, err := io.Copy(output, response.Body)
	if err != nil {
		return err
	}

	w.Write([]byte(string(n) + " bytes downloaded."))
	return nil
}

func UploadFromFileToUrl(url, fileName string, w io.Writer) error {
	if url == "" {
		return fmt.Errorf("To upload from a file to a url, I need a URL.")
	}
	w.Write([]byte("Uploading " + url + " to " + fileName))

	input, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer input.Close()

	request, err := http.NewRequest("POST", url, input)
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	_, err = io.Copy(w, response.Body)
	if err != nil {
		return err
	}

	return nil
}

func IPFSBaseUrl() string {
	host := "http://localhost:8080"
	if os.Getenv("ERIS_CLI_CONTAINER") == "true" {
		host = "http://ipfs:8080"
	}
	return host + "/ipfs/"
}
