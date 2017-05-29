package perform

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/monax/monax/log"
	"github.com/monax/compilers/definitions"
)

// send an http request and wait for the response
func requestResponse(req *definitions.Request, URL string) (*Response, error) {
	// make request
	reqJ, err := json.Marshal(req)
	if err != nil {
		log.Errorln("failed to marshal req obj", err)
		return nil, err
	}
	httpreq, err := http.NewRequest("POST", URL, bytes.NewBuffer(reqJ))
	if err != nil {
		log.Errorln("failed to compose request:", err)
		return nil, err
	}
	httpreq.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(httpreq)
	if err != nil {
		log.Errorln("failed to send HTTP request", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode > 300 {
		return nil, fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	respJ := new(Response)
	// read in response body
	body, err := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(body, respJ)
	if err != nil {
		log.Errorln("failed to unmarshal", err)
		return nil, err
	}
	return respJ, nil
}

// send an http request and wait for the response
func requestBinaryResponse(req *definitions.BinaryRequest, URL string) (*BinaryResponse, error) {
	// make request
	reqJ, err := json.Marshal(req)
	if err != nil {
		log.Errorln("failed to marshal req obj", err)
		return nil, err
	}
	httpreq, err := http.NewRequest("POST", URL, bytes.NewBuffer(reqJ))
	if err != nil {
		log.Errorln("failed to compose request:", err)
		return nil, err
	}
	httpreq.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(httpreq)
	if err != nil {
		log.Errorln("failed to send HTTP request", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode > 300 {
		return nil, fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	respJ := new(BinaryResponse)
	// read in response body
	body, err := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(body, respJ)
	if err != nil {
		log.Errorln("failed to unmarshal", err)
		return nil, err
	}
	return respJ, nil
}
