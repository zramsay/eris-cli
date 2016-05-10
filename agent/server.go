package agent

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/util"

	log "github.com/Sirupsen/logrus"
	"github.com/rs/cors"
)

func StartAgent(do *definitions.Do) error {
	mux := http.NewServeMux()
	mux.Handle("/chains", agentHandler(ListChains))
	mux.Handle("/download", agentHandler(DownloadAgent))
	mux.Handle("/install", agentHandler(InstallAgent))
	fmt.Println(`Starting agent on localhost:17552

Available endpoints are:
/chains (GET)

/download (POST) 
  => /download?groupId=<abc>&bundleId=<def>&version=<1.0.2>&hash=<ipfs>

/install (POST)
  => /install?groupId=<abc>&bundleId=<def>&version=<1.0.2>&hash=<ipfs>&chainName=<alice>&address=<addr>"
`)

	// cors.Default() sets up the middleware with default options being
	// all origins accepted with simple methods (GET, POST).
	// See https://github.com/rs/cors
	handler := cors.Default().Handler(mux)

	if err := http.ListenAndServe(":17552", handler); err != nil {
		return fmt.Errorf("error starting agent: %v", err)
	}

	return nil
}

/*func StopAgent(do *definitions.Do) error {
	fmt.Println("Gracefully shutting down agent")
	return nil
}*/

type agentError struct {
	Error   error
	Message string
	Code    int
}

type agentHandler func(http.ResponseWriter, *http.Request) *agentError

func (endpoint agentHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := endpoint(w, r); err != nil {
		http.Error(w, fmt.Sprintf("%s %v", err.Message, err.Error), err.Code)
	}
}

// TODO move to errors packags
var (
	ErrorDownloadingBundle       = "error downloading bundle:"
	ErrorReadingTarball          = "error reading tarball:"
	ErrorDeployingContractBundle = "error deploying contract bundle:"
	ErrorParsingURL              = "error parsing url:"
	ErrorCheckingIPFShash        = "error checking ipfs hash:"
	ErrorReadingEPMjson          = "error reading epm.json file:"
)

/* status codes we use
StatusBadRequest = 400
StatusNotFound = 404
StatusInternalServerError = 500

*/

// todo catch an error??
func ListChains(w http.ResponseWriter, r *http.Request) *agentError {
	if r.Method == "GET" {
		var deets []*util.Details

		deets = util.ErisContainersByType(definitions.TypeChain, true)
		//if len(deets) == 0 {
		//} return proper thing
		names := make([]string, len(deets))
		for i, chainName := range deets {
			names[i] = chainName.ShortName
		}

		chainz := []string{}
		lenChainz := len(names) - 1
		for i, fmtName := range names {
			if i == 0 {
				chainz = append(chainz, "[")
			}
			if i == lenChainz {
				chainz = append(chainz, fmt.Sprintf("{ name: '%s' }]", fmtName))
				break
			}
			chainz = append(chainz, fmt.Sprintf("{ name: '%s' }, ", fmtName))
		}
		w.Write([]byte(strings.Join(chainz, "")))
	}
	return nil
}

func DownloadAgent(w http.ResponseWriter, r *http.Request) *agentError {
	if r.Method == "POST" {
		log.Warn("Receiving request to download a contract bundle")
		reqArgs := []string{"groupId", "bundleId", "version", "hash"}
		params, err := ParseURL(reqArgs, fmt.Sprintf("%s", r.URL))
		if err != nil {
			return &agentError{err, ErrorParsingURL, 400}
		}

		installPath := SetTarballPath(params)
		whichHash, err := checkHash(params["hash"])
		if err != nil {
			return &agentError{err, ErrorCheckingIPFShash, 400}
		}

		if whichHash == "tarball" {
			tarBody, err := ioutil.ReadAll(r.Body)
			if err != nil {
				return &agentError{err, ErrorReadingTarball, 400}
			}

			if err := downloadBundleFromTarball(tarBody, installPath, params["hash"]); err != nil {
				return &agentError{err, ErrorDownloadingBundle, 500}
			}

		} else if whichHash == "ipfs-hash" { // not directly tarball, get from ipfs
			if err := downloadBundleFromIPFS(params); err != nil {
				return &agentError{err, ErrorDownloadingBundle, 500}
			}
		}
	}
	return nil
}

func InstallAgent(w http.ResponseWriter, r *http.Request) *agentError {
	if r.Method == "POST" {
		log.Warn("Receiving request to download and deploy a contract bundle")

		// parse response into various components
		// required to pull tarball, unpack on path
		// and deploy to running chain
		reqArgs := []string{"groupId", "bundleId", "version", "hash", "chainName", "address"}
		params, err := ParseURL(reqArgs, fmt.Sprintf("%s", r.URL))
		if err != nil {
			return &agentError{err, ErrorParsingURL, 400}
		}

		// ensure chain to deploy on is running
		// might want to perform some other checks ... ?
		if !IsChainRunning(params["chainName"]) {
			return &agentError{nil, "chain name provided is not running", 404}
		}

		installPath := SetTarballPath(params)
		whichHash, err := checkHash(params["hash"])
		if err != nil {
			return &agentError{err, ErrorCheckingIPFShash, 400}
		}

		if whichHash == "tarball" {
			tarBody, err := ioutil.ReadAll(r.Body)
			if err != nil {
				return &agentError{err, ErrorReadingTarball, 400}
			}

			if err := downloadBundleFromTarball(tarBody, installPath, params["hash"]); err != nil {
				return &agentError{err, ErrorDownloadingBundle, 500}
			}

		} else if whichHash == "ipfs-hash" {
			if err := downloadBundleFromIPFS(params); err != nil {
				return &agentError{err, ErrorDownloadingBundle, 500}
			}
		}

		// chain is running
		// contract bundle unbundled
		// time to deploy
		if err := DeployContractBundle(installPath, params["chainName"], params["address"]); err != nil {
			return &agentError{err, ErrorDeployingContractBundle, 403}
			// TODO reap bad addr error => func AuthenticateUser()
		}

		epmJSON := filepath.Join(installPath, "epm.json")
		epmByte, err := ioutil.ReadFile(epmJSON)
		if err != nil {
			return &agentError{err, ErrorReadingEPMjson, 500}
		}
		w.Write(epmByte)
	}
	return nil
}
