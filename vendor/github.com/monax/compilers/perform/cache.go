package perform

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"

	"github.com/monax/compilers/definitions"
)

// check/cache all includes, hash the code, return whether or not there was a full cache hit
func CheckCached(includes map[string]*definitions.IncludedFiles, lang string) bool {
	cached := true
	for name, metadata := range includes {
		hashPath := path.Join(definitions.Languages[lang].CacheDir, name)
		if _, scriptErr := os.Stat(hashPath); os.IsNotExist(scriptErr) {
			cached = false
			break
		}
		for _, object := range metadata.ObjectNames {
			objectFile := path.Join(hashPath, object+".json")
			if _, objErr := os.Stat(objectFile); objErr != nil {
				cached = false
				break
			}
		}
		if cached == false {
			break
		}
	}

	return cached
}

// return cached byte code as a response
func CachedResponse(includes map[string]*definitions.IncludedFiles, lang string) (*Response, error) {

	var resp *Response
	var respItemArray []ResponseItem
	for name, metadata := range includes {
		dir := path.Join(definitions.Languages[lang].CacheDir, name)
		for _, object := range metadata.ObjectNames {
			jsonBytes, err := ioutil.ReadFile(path.Join(dir, object+".json"))
			if err != nil {
				return nil, err
			}
			respItem := &ResponseItem{}
			err = json.Unmarshal(jsonBytes, respItem)
			if err != nil {
				return nil, err
			}
			respItemArray = append(respItemArray, *respItem)
		}
	}
	resp = &Response{
		Objects: respItemArray,
		Warning: "",
		Error:   "",
	}

	return resp, nil
}

// cache ABI and Binary to
func CacheResult(object ResponseItem, cacheLocation, warning, version, errorString string) error {
	os.Chdir(cacheLocation)
	fullResponse := Response{[]ResponseItem{object}, warning, version, errorString}
	cachedObject, err := json.Marshal(fullResponse)
	ioutil.WriteFile(object.Objectname+".json", []byte(cachedObject), 0644)
	return err
}
