package agent

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/files"
	"github.com/eris-ltd/eris-cli/pkgs"
	"github.com/eris-ltd/eris-cli/util"

	log "github.com/Sirupsen/logrus"
	"github.com/eris-ltd/common/go/common"
	"github.com/rs/cors"
)

func StartAgent(do *definitions.Do) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/chains", ListChains)
	mux.HandleFunc("/download", DownloadAgent)
	mux.HandleFunc("/install", InstallAgent)
	fmt.Println("Starting agent on localhost:17552")

	// cors.Default() sets up the middleware with default options being
	// all origins accepted with simple methods (GET, POST).
	// See https://github.com/rs/cors
	handler := cors.Default().Handler(mux)

	if err := http.ListenAndServe(":17552", handler); err != nil {
		return fmt.Errorf("Error starting agent: %v", err)
	}

	return nil
}

/*func StopAgent(do *definitions.Do) error {
	fmt.Println("Gracefully shutting down agent")
	return nil
}*/

func ListChains(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		var deets []*util.Details

		deets = util.ErisContainersByType(definitions.TypeChain, true)
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
}

func DownloadAgent(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		log.Warn("Receiving request to download a contract bundle")
		reqArgs := []string{"groupId", "bundleId", "version", "hash"}
		params, err := ParseURL(reqArgs, fmt.Sprintf("%s", r.URL))
		if err != nil {
			errMsg := fmt.Sprintf("error parsing url: %v", err)
			http.Error(w, errMsg, http.StatusBadRequest)
			return
		}

		installPath := SetTarballPath(params)
		whichHash, err := checkHash(params["hash"])
		if err != nil {
			return
		}

		if whichHash == "tarball" {
			tarBody, err := ioutil.ReadAll(r.Body)
			if err != nil {
				errMsg := fmt.Sprintf("error reading tarball body: %v", err)
				http.Error(w, errMsg, http.StatusInternalServerError)
				return
			}

			if err := downloadBundleFromTarball(tarBody, installPath, params["hash"]); err != nil {
				errMsg := fmt.Sprintf("error downloading bundle: %v", err)
				http.Error(w, errMsg, http.StatusInternalServerError)
				return
			}

		} else if whichHash == "ipfs-hash" { // not directly tarball, get from ipfs
			if err := downloadBundleFromIPFS(params); err != nil {
				errMsg := fmt.Sprintf("error downloading bundle: %v", err)
				http.Error(w, errMsg, http.StatusInternalServerError)
				return
			}
		}
	}
}

func InstallAgent(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		log.Warn("Receiving request to download and deploy a contract bundle")

		// parse response into various components
		// required to pull tarball, unpack on path
		// and deploy to running chain
		reqArgs := []string{"groupId", "bundleId", "version", "hash", "chainName", "address"}
		params, err := ParseURL(reqArgs, fmt.Sprintf("%s", r.URL))
		if err != nil {
			errMsg := fmt.Sprintf("error parsing url: %v", err)
			http.Error(w, errMsg, http.StatusBadRequest)
			return
		}

		// ensure chain to deploy on is running
		// might want to perform some other checks ... ?
		if !IsChainRunning(params["chainName"]) {
			http.Error(w, `chain name provided is not running`, http.StatusNotFound)
			return
		}

		installPath := SetTarballPath(params)
		whichHash, err := checkHash(params["hash"])
		if err != nil {
			errMsg := fmt.Sprintf("error checking hash: %v", err)
			http.Error(w, errMsg, http.StatusBadRequest)
			return
		}

		if whichHash == "tarball" {
			tarBody, err := ioutil.ReadAll(r.Body)
			if err != nil {
				errMsg := fmt.Sprintf("error reading tarball body: %v", err)
				http.Error(w, errMsg, http.StatusInternalServerError)
				return
			}

			if err := downloadBundleFromTarball(tarBody, installPath, params["hash"]); err != nil {
				errMsg := fmt.Sprintf("error downloading bundle: %v", err)
				http.Error(w, errMsg, http.StatusBadRequest)
				return
			}

		} else if whichHash == "ipfs-hash" {
			if err := downloadBundleFromIPFS(params); err != nil {
				errMsg := fmt.Sprintf("error downloading bundle: %v", err)
				http.Error(w, errMsg, http.StatusInternalServerError)
				return
			}
		}

		// chain is running
		// contract bundle unbundled
		// time to deploy
		if err := DeployContractBundle(installPath, params["chainName"], params["address"]); err != nil {
			errMsg := fmt.Sprintf("error deploying contract bundle: %v", err)
			http.Error(w, errMsg, http.StatusForbidden)
			// TODO reap bad addr error => func AuthenticateUser()
			return
		}

		epmJSON := filepath.Join(installPath, "epm.json")
		epmByte, err := ioutil.ReadFile(epmJSON)
		if err != nil {
			errMsg := fmt.Sprintf("error reading epm.json file: %v", err)
			http.Error(w, errMsg, http.StatusInternalServerError)
		}
		w.Write(epmByte)
	}
}

func checkHash(hash string) (string, error) {
	splitHash := strings.Split(hash, ".")
	if len(splitHash) >= 3 { // somefile.tar.gz for example
		lenTAR := len(splitHash) - 2
		lenGZ := len(splitHash) - 1
		//probably not ipfs hash & probably tar ball
		if splitHash[lenTAR] == "tar" && splitHash[lenGZ] == "gz" {
			log.Debug("hash provided appears to be a tarball")
			return "tarball", nil
		}
	}

	if len(hash) == 46 { // ipfs hash
		log.Debug("hash provided appears to be an IPFS hash")
		return "ipfs-hash", nil
	}
	return "", fmt.Errorf("unable to decipher hash")
}

func downloadBundleFromIPFS(params map[string]string) error {

	installPath := SetTarballPath(params)

	tarBallPath, err := GetTarballFromIPFS(params["hash"], installPath)
	if err != nil {
		//http.Error(w, `error getting from IPFS`, http.StatusInternalServerError)
		return err
	}

	if err = UnpackTarball(tarBallPath, installPath); err != nil {
		//http.Error(w, `error unpacking tarball`, http.StatusInternalServerError)
		return err
	}

	return nil
}

func downloadBundleFromTarball(body []byte, installPath, fileName string) error {
	if err := os.MkdirAll(installPath, 0777); err != nil {
		return err
	}

	tarBallPath := filepath.Join(installPath, fileName)
	if err := ioutil.WriteFile(tarBallPath, body, 0777); err != nil {
		return err
	}

	if err := UnpackTarball(tarBallPath, installPath); err != nil {
		return err
	}
	return nil
}

// takes URL request for bundle install &
// returns a map of the things we need for installation
// assume it comes in as r.URL.Path[1:]
func ParseURL(requiredArguments []string, url string) (map[string]string, error) {
	reqArgs := requiredArguments

	payload := strings.Split(url, "?")[1]
	infos := strings.Split(payload, "&")
	parsedURL := make(map[string]string, len(infos))

	for _, field := range infos {
		info := strings.Split(field, "=")
		parsedURL[info[0]] = info[1] // this is not very resilient
	}

	if len(parsedURL) != len(reqArgs) {
		return nil, fmt.Errorf("Wrong number of arguments:\n%v\nThese %v arguments are required:\n%s", parsedURL, len(reqArgs), strings.Join(reqArgs, ", "))
	}

	for _, arg := range reqArgs {
		if parsedURL[arg] == "" {
			return nil, fmt.Errorf("Missing field or bad argument name:\n%v\nThese fields cannot be empty:\n%s", parsedURL, strings.Join(reqArgs, ", "))
		}
	}

	return parsedURL, nil
}

func IsChainRunning(chainName string) bool {
	return util.IsChain(chainName, true)
}

// returns path to downloaded tarball
func GetTarballFromIPFS(hash, installPath string) (string, error) {
	if !util.DoesDirExist(installPath) {
		if err := os.MkdirAll(installPath, 0777); err != nil {
			return "", err
		}
	}
	do := definitions.NowDo()
	do.Name = hash

	do.Path = filepath.Join(installPath, fmt.Sprintf("%s%s", hash, ".tar.gz"))
	if err := files.GetFiles(do); err != nil {
		return "error", err
	}

	return do.Path, nil
}

// for consistency in InstallAgent()
func UnpackTarball(tarBallPath, installPath string) error {
	return util.UnpackTarball(tarBallPath, installPath)
}

func DeployContractBundle(path, chainName, address string) error {

	doRun := definitions.NowDo()
	doRun.Path = path
	doRun.EPMConfigFile = filepath.Join(path, "epm.yaml")
	doRun.PackagePath = path
	doRun.ABIPath = path
	doRun.ChainName = chainName
	doRun.DefaultAddr = address

	if err := pkgs.RunPackage(doRun); err != nil {
		return err
	}
	log.Warn(doRun.Result)

	return nil
}

func SetTarballPath(bundleInfo map[string]string) string {
	groupID := strings.Replace(bundleInfo["groupId"], ".", "/", 1)
	bundleID := bundleInfo["bundleId"]
	version := bundleInfo["version"]

	return filepath.Join(common.BundlesPath, groupID, bundleID, version)
}

// ensure user is valid
// by querying the accounts chain
// TODO implement. Blocking on lack of details
//if !AuthenticateUser(params["user"]) {
//	w.Write([]byte(fmt.Sprintf("permissioned denied: %v\n", err)))
//	return
//}
