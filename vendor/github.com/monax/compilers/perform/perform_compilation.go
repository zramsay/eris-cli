package perform

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/monax/compilers/definitions"
	"github.com/monax/compilers/util"

	"github.com/monax/monax/config"
	"github.com/monax/monax/log"
)

type Response struct {
	Objects []ResponseItem `json:"objects"`
	Warning string         `json:"warning"`
	Version string         `json:"version"`
	Error   string         `json:"error"`
}

type BinaryResponse struct {
	Binary string `json:"binary"`
	Error  string `json:"error"`
}

// Compile response object
type ResponseItem struct {
	Objectname string `json:"objectname"`
	Bytecode   string `json:"bytecode"`
	ABI        string `json:"abi"` // json encoded
}

func (resp Response) CacheNewResponse(req definitions.Request) {
	objects := resp.Objects
	//log.Debug(objects)
	cacheLocation := definitions.Languages[req.Language].CacheDir
	cur, _ := os.Getwd()
	os.Chdir(cacheLocation)
	defer func() {
		os.Chdir(cur)
	}()
	for fileDir, metadata := range req.Includes {
		dir := path.Join(cacheLocation, strings.TrimRight(fileDir, "."+req.Language))
		os.MkdirAll(dir, 0700)
		objectNames := metadata.ObjectNames
		for _, name := range objectNames {
			for _, object := range objects {
				if object.Objectname == name {
					//log.WithField("=>", resp.Objects).Debug("Response objects over the loop")
					CacheResult(object, dir, resp.Warning, resp.Version, resp.Error)
					break
				}
			}
		}
	}
}

func linkBinaries(req *definitions.BinaryRequest) *BinaryResponse {
	// purely for solidity and solidity alone as this is soon to be deprecated.
	if req.Libraries == "" {
		return &BinaryResponse{
			Binary: req.BinaryFile,
			Error:  "",
		}
	}

	buf := bytes.NewBufferString(req.BinaryFile)
	var output bytes.Buffer
	var stderr bytes.Buffer
	linkCmd := exec.Command("solc", "--link", "--libraries", req.Libraries)
	linkCmd.Stdin = buf
	linkCmd.Stderr = &stderr
	linkCmd.Stdout = &output
	linkCmd.Start()
	linkCmd.Wait()

	return &BinaryResponse{
		Binary: strings.TrimSpace(output.String()),
		Error:  stderr.String(),
	}
}

func RequestBinaryLinkage(url string, file string, libraries string) (*BinaryResponse, error) {
	//Create Binary Request, send it off
	code, err := ioutil.ReadFile(file)
	if err != nil {
		return &BinaryResponse{}, err
	}
	request := &definitions.BinaryRequest{
		BinaryFile: string(code),
		Libraries:  libraries,
	}
	return requestBinaryResponse(request, url)
}

//todo: Might also need to add in a map of library names to addrs
func RequestCompile(url string, file string, optimize bool, libraries string) (*Response, error) {
	config.InitMonaxDir()
	request, err := CreateRequest(file, libraries, optimize)
	if err != nil {
		return nil, err
	}
	//todo: check server for newer version of same files...
	// go through all includes, check if they have changed
	cached := CheckCached(request.Includes, request.Language)

	log.WithField("cached?", cached).Debug("Cached Item(s)")

	/*for k, v := range request.Includes {
		log.WithFields(log.Fields{
			"k": k,
			"v": string(v.Script),
		}).Debug("check request loop")
	}*/

	var resp *Response
	// if everything is cached, no need for request
	if cached {
		// TODO: need to return all contracts/libs tied to the original src file
		resp, err = CachedResponse(request.Includes, request.Language)
		if err != nil {
			return nil, err
		}
	} else {
		log.Debug("Could not find cached object, compiling...")
		if url == "" {
			resp = compile(request)
		} else {
			resp, err = requestResponse(request, url)
			if err != nil {
				return nil, err
			}
		}
		resp.CacheNewResponse(*request)
	}

	PrintResponse(*resp, false)

	return resp, nil
}

// Compile takes a dir and some code, replaces all includes, checks cache, compiles, caches
func compile(req *definitions.Request) *Response {

	if _, ok := definitions.Languages[req.Language]; !ok {
		return compilerResponse("", "", "", "", "", fmt.Errorf("No script provided"))
	}

	lang := definitions.Languages[req.Language]

	includes := []string{}
	currentDir, _ := os.Getwd()
	defer os.Chdir(currentDir)

	for k, v := range req.Includes {
		os.Chdir(lang.CacheDir)
		file, err := util.CreateTemporaryFile(k, v.Script)
		if err != nil {
			return compilerResponse("", "", "", "", "", err)
		}
		defer os.Remove(file.Name())
		includes = append(includes, file.Name())
		log.WithField("Filepath of include: ", file.Name()).Debug("To Cache")
	}

	libsFile, err := util.CreateTemporaryFile("monax-libs", []byte(req.Libraries))
	if err != nil {
		return compilerResponse("", "", "", "", "", err)
	}
	defer os.Remove(libsFile.Name())
	command := lang.Cmd(includes, libsFile.Name(), req.Optimize)
	log.WithField("Command: ", command).Debug("Command Input")
	output, err := runCommand(command...)

	var warning string
	jsonBeginsCertainly := strings.Index(output, `{"contracts":`)

	if jsonBeginsCertainly > 0 {
		warning = output[:jsonBeginsCertainly]
		output = output[jsonBeginsCertainly:]
	}

	//cleanup
	log.WithField("=>", output).Debug("Output from command: ")
	if err != nil {
		for _, str := range includes {
			output = strings.Replace(output, str, req.FileReplacement[str], -1)
		}
		log.WithFields(log.Fields{
			"err":      err,
			"command":  command,
			"response": output,
		}).Debug("Could not compile")
		return compilerResponse("", "", "", "", "", fmt.Errorf("%v", output))
	}

	solcResp := definitions.BlankSolcResponse()

	//todo: provide unmarshalling for serpent and lll
	log.WithField("Json: ", output).Debug("Command Output")
	err = json.Unmarshal([]byte(output), solcResp)
	if err != nil {
		log.Debug("Could not unmarshal json")
		return compilerResponse("", "", "", "", "", err)
	}
	respItemArray := make([]ResponseItem, 0)

	for contract, item := range solcResp.Contracts {
		respItem := ResponseItem{
			Objectname: strings.TrimSpace(contract),
			Bytecode:   strings.TrimSpace(item.Bin),
			ABI:        strings.TrimSpace(item.Abi),
		}
		respItemArray = append(respItemArray, respItem)
	}

	for _, re := range respItemArray {
		log.WithFields(log.Fields{
			"name": re.Objectname,
			"bin":  re.Bytecode,
			"abi":  re.ABI,
		}).Debug("Response formulated")
	}

	return &Response{
		Objects: respItemArray,
		Warning: warning,
		Error:   "",
	}
}

func runCommand(tokens ...string) (string, error) {
	cmd := tokens[0]
	args := tokens[1:]
	shellCmd := exec.Command(cmd, args...)
	output, err := shellCmd.CombinedOutput()
	s := strings.TrimSpace(string(output))
	return s, err
}

func CreateRequest(file string, libraries string, optimize bool) (*definitions.Request, error) {
	var includes = make(map[string]*definitions.IncludedFiles)

	//maps hashes to original file name
	var hashFileReplacement = make(map[string]string)
	language, err := util.LangFromFile(file)
	if err != nil {
		return &definitions.Request{}, err
	}
	compiler := &definitions.Compiler{
		Config: definitions.Languages[language],
		Lang:   language,
	}
	code, err := ioutil.ReadFile(file)
	if err != nil {
		return &definitions.Request{}, err
	}
	dir := path.Dir(file)
	//log.Debug("Before parsing includes =>\n\n%s", string(code))
	code, err = compiler.ReplaceIncludes(code, dir, file, includes, hashFileReplacement)
	if err != nil {
		return &definitions.Request{}, err
	}

	return compiler.CompilerRequest(file, includes, libraries, optimize, hashFileReplacement), nil
}

// New response object from bytecode and an error
func compilerResponse(objectname, bytecode, abi, warning, version string, err error) *Response {
	e := ""
	if err != nil {
		e = err.Error()
	}

	respItem := ResponseItem{
		Objectname: objectname,
		Bytecode:   bytecode,
		ABI:        abi}

	respItemArray := make([]ResponseItem, 1)
	respItemArray[0] = respItem

	return &Response{
		Objects: respItemArray,
		Warning: warning,
		Version: version,
		Error:   e,
	}
}

func PrintResponse(resp Response, cli bool) {
	if resp.Error != "" {
		log.Warn(resp.Error)
	} else {
		for _, r := range resp.Objects {
			message := log.WithFields((log.Fields{
				"name": r.Objectname,
				"bin":  r.Bytecode,
				"abi":  r.ABI,
			}))
			if cli {
				message.Warn("Response")
			} else {
				message.Info("Response")
			}
		}
	}
}
