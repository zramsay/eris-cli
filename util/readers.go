package util

import (
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
	w.Write([]byte("GETing file from IPFS. Hash -> " + hash + "\n"))
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

	response, err := http.Get(url)
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

func IPFSBaseUrl() string {
	host := "http://localhost:8080"
	if os.Getenv("ERIS_CLI_CONTAINER") == "true" {
		host = "http://ipfs:8080"
	}
	return host + "/ipfs/"
}
