package util

import (
  "bytes"
  "fmt"
  "io"
  "io/ioutil"
  "net/http"
  "os"
  "strings"
)

func DownloadFromGithub(org, repo, branch, path, fileName string, w io.Writer) error {
  url := "https://rawgit.com/" + strings.Join([]string{org, repo, branch, path}, "/")
  w.Write([]byte("Will download from url -> " + url))
  return DownloadFromUrl(url, fileName, w)
}

func DownloadFromUrl(url, fileName string, w io.Writer) error {
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

func GetFromIPFS(hash string, w io.Writer) error {
    req, err  := http.NewRequest("GET", IPFSBaseUrl() + hash, bytes.NewBuffer([]byte{}))
    client    := &http.Client{}
    resp, err := client.Do(req)

    if err != nil {
      return err
    }
    defer resp.Body.Close()

    w.Write([]byte("response Status:" + resp.Status))
    // w.Write([]byte("response Headers:" + resp.Header))

    body, e := ioutil.ReadAll(resp.Body)
    if e != nil {
      return e
    }

    if (len(string(body)) <= 10000) {
      w.Write([]byte("REQUEST: " + string(body)))
    } else {
      toPrint := body[:10000]
      w.Write([]byte("REQUEST: " + string(toPrint) + " ...{truncated}"))
    }
    fmt.Println(string(body))

  return nil
}

func IPFSBaseUrl() string {
  host := "http://localhost:8080"
  // host = "http://ipfs:8080"
  return host + "/ipfs/"
}

// func (api *IpfsApi) get(hash string) (string,error) {


//     return nil
// }

// func (api *IpfsApi) post(data []byte) (string,error){
//   fmt.Println("POSTing")
//     // var jsonStr = []byte(`{"title":"Buy cheese and bread for breakfast."}`)
//     req, err := http.NewRequest("POST", BaseUrl(), bytes.NewBuffer(data))
//     if err != nil {
//       fmt.Println("Error POSTing: " + err.Error())
//       return "", err
//     }
//     req.Header.Set("Content-Type", "application/json")

//     resp, err2 := api.client.Do(req)

//     if err2 != nil {
//       fmt.Println("Error finalizing POST: " + err2.Error())
//         return "",err
//     }
//     defer resp.Body.Close()

//     fmt.Println("response Status:", resp.Status)
//     fmt.Println("response Headers:", resp.Header)
//     body, _ := ioutil.ReadAll(resp.Body)
//     if (len(string(body)) <= 10000) {
//       fmt.Println("REQUEST: " + string(body))
//     } else {
//       toPrint := body[:10000]
//       fmt.Println("REQUEST: " + string(toPrint) + " ...{truncated}")
//     }
//     // fmt.Println("response Body:", string(body))
//     hash, ok := resp.Header["Ipfs-Hash"]
//     if !ok || hash[0] == "" {
//       // Should not happen
//       return "",fmt.Errorf("No hash returned");
//     }
//     fmt.Println("HASH: " + hash[0])
//     return hash[0],nil
// }