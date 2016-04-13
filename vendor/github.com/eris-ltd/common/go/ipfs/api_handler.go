package ipfs

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

//returns []byte to let each command make its own struct for the response
//but handles the errs in here
func PostAPICall(url, fileHash string, w io.Writer) ([]byte, error) {
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
		return []byte(""), fmt.Errorf("A timeout occured while trying to reach IPFS. Run `eris files cache [hash], wait 5-10 seconds, then run `eris files [cmd] [hash]`")
	}
	return body, nil
}
