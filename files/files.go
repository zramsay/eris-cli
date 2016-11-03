package files

import (
	"bytes"
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/eris-ltd/eris-cli/config"
	"github.com/eris-ltd/eris-cli/data"
	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/log"
	"github.com/eris-ltd/eris-cli/services"
	"github.com/eris-ltd/eris-cli/util"
)

func GetFiles(do *definitions.Do) error {
	if err := EnsureIPFSrunning(); err != nil {
		return err
	}

	// where Object is a directory already added recursively to ipfs
	dirBool, err := isHashAnObject(do.Hash)
	if err != nil {
		return err
	}

	if dirBool {
		log.WithFields(log.Fields{
			"hash": do.Hash,
			"path": do.Path,
		}).Warn("Getting a directory")
		buf, err := importDirectory(do)
		if err != nil {
			return err
		}
		log.Warn("Directory object getted succesfully.")
		log.Warn(strings.TrimSpace(buf.String()))
	} else {
		if err := importFile(do.Hash, do.Path, do.IpfsPort); err != nil {
			return err
		}
	}
	return nil
}

func PutFiles(do *definitions.Do) (string, error) {
	if err := EnsureIPFSrunning(); err != nil {
		return "", err
	}

	if err := checkGatewayFlag(do); err != nil {
		return "", err
	}

	// Check if do.Name is dir or file.
	f, err := os.Stat(do.Name)
	if err != nil {
		return "", err
	}

	if f.IsDir() {
		log.WithField("dir", do.Name).Warn("Adding contents of a directory")
		buf, err := exportDirectory(do)
		if err != nil {
			return "", err
		}
		log.Warn("Directory object added succesfully")
		log.Warn(strings.TrimSpace(buf.String()))
	} else {
		hash, err := exportFile(do.Name, do.Gateway, do.IpfsPort)
		if err != nil {
			return "", err
		}
		return hash, nil
	}
	return "", nil
}

func exportDirectory(do *definitions.Do) (*bytes.Buffer, error) {
	// path to dir on host
	do.Source = do.Name
	do.Destination = filepath.Join(config.ErisContainerRoot, "scratch", "data", do.Source)
	do.Name = "ipfs"

	do.Operations.Args = nil
	do.Operations.PublishAllPorts = true
	if err := data.ImportData(do); err != nil {
		return nil, err
	}

	ip := new(bytes.Buffer)
	config.Global.Writer = ip

	do.Operations.Interactive = false
	do.Operations.PublishAllPorts = true
	do.Operations.Args = []string{"NetworkSettings.IPAddress"}

	if err := services.InspectService(do); err != nil {
		return nil, err
	}
	api := fmt.Sprintf("/ip4/%s/tcp/5001", strings.TrimSpace(ip.String()))

	argumentsAdd := []string{"ipfs", "add", "-r", do.Destination, "--api", api}

	buf, err := services.ExecHandler("ipfs", argumentsAdd)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

func importDirectory(do *definitions.Do) (*bytes.Buffer, error) {
	hash := do.Hash

	ip := new(bytes.Buffer)
	config.Global.Writer = ip

	do.Name = "ipfs"
	do.Operations.Interactive = false
	do.Operations.PublishAllPorts = true
	do.Operations.Args = []string{"NetworkSettings.IPAddress"}

	if err := services.InspectService(do); err != nil {
		return nil, err
	}
	api := fmt.Sprintf("/ip4/%s/tcp/5001", strings.TrimSpace(ip.String()))

	argumentsGet := []string{"ipfs", "get", hash, "--api", api}

	buf, err := services.ExecHandler("ipfs", argumentsGet)
	if err != nil {
		return nil, err
	}

	do.Destination = do.Path
	do.Source = path.Join(config.ErisContainerRoot, hash)
	do.Operations.Args = nil
	do.Operations.PublishAllPorts = false
	if err := data.ExportData(do); err != nil {
		return nil, err
	}

	_, err = os.Getwd()
	if err != nil {
		return nil, err
	}
	theDir := path.Join(do.Destination, hash)
	newDir := do.Destination

	if err := data.MoveOutOfDirAndRmDir(theDir, newDir); err != nil {
		return nil, err
	}

	return buf, nil

}
func PinFiles(do *definitions.Do) (string, error) {
	if err := EnsureIPFSrunning(); err != nil {
		return "", err
	}
	log.WithFields(log.Fields{
		"file": do.Name,
		"path": do.Path,
	}).Debug("Pinning a file")
	hash, err := pinFile(do.Name)
	if err != nil {
		return "", err
	}
	return hash, nil
}

func CatFiles(do *definitions.Do) (string, error) {
	if err := EnsureIPFSrunning(); err != nil {
		return "", err
	}

	log.WithFields(log.Fields{
		"file": do.Name,
		"path": do.Path,
	}).Debug("Dumping the contents of a file")
	hash, err := catFile(do.Name)
	if err != nil {
		return "", err
	}
	return hash, nil
}

func ListFiles(do *definitions.Do) (string, error) {
	if err := EnsureIPFSrunning(); err != nil {
		return "", err
	}

	log.WithFields(log.Fields{
		"file": do.Name,
		"path": do.Path,
	}).Debug("Listing an object")
	hash, err := listFile(do.Name)
	if err != nil {
		return "", err
	}
	return hash, nil
}

func ManagePinned(do *definitions.Do) (string, error) {
	if err := EnsureIPFSrunning(); err != nil {
		return "", err
	}
	if do.Rm && do.Hash != "" {
		return "", fmt.Errorf("Either remove a file by hash or all of them")
	}

	if do.Rm {
		log.Info("Removing all cached files")
		hashes, err := rmAllPinned()
		if err != nil {
			return "", err
		}
		return hashes, nil
	} else if do.Hash != "" {
		log.WithField("hash", do.Hash).Info("Removing from cache")
		hashes, err := rmPinnedByHash(do.Hash)
		if err != nil {
			return "", err
		}
		return hashes, nil
	}
	log.Debug("Listing files pinned locally")
	hash, err := listPinned()
	if err != nil {
		return "", err
	}
	return hash, nil
}

func importFile(hash, fileName, port string) error {
	log.WithFields(log.Fields{
		"from hash": hash,
		"to path":   fileName,
	}).Debug("Importing a file")

	return util.GetFromIPFS(hash, fileName, "", port)
}

func exportFile(fileName, gateway, port string) (string, error) {
	log.WithFields(log.Fields{
		"file":    fileName,
		"gateway": gateway,
	}).Debug("Adding a file")

	return util.SendToIPFS(fileName, gateway, port)
}

func pinFile(fileHash string) (string, error) {
	return util.PinToIPFS(fileHash)
}

func catFile(fileHash string) (string, error) {
	return util.CatFromIPFS(fileHash)
}

func listFile(objectHash string) (string, error) {
	hash, err := util.ListFromIPFS(objectHash)

	if err != nil {
		if fmt.Sprintf("%v", err) != "EOF" {
			return "", err
		} else {
			return hash, nil
		}
	}
	return hash, nil
}

func listPinned() (string, error) {
	return util.ListPinnedFromIPFS()
}

func rmAllPinned() (string, error) {
	hashList, err := listPinned()
	if err != nil {
		return "", err
	}

	hashArray := strings.Split(hashList, "\n")
	result := make([]string, len(hashArray))
	for i, hash := range hashArray {
		result[i], err = rmPinnedByHash(hash)
		if err != nil {
			return "", err
		}
	}
	hashes := strings.Join(result, "\n")
	return hashes, nil
}

func rmPinnedByHash(hash string) (string, error) {
	return util.RemovePinnedFromIPFS(hash)
}

func EnsureIPFSrunning() error {
	do := definitions.NowDo()
	do.Name = "ipfs"
	if err := services.EnsureRunning(do); err != nil {
		return fmt.Errorf("Failed to ensure IPFS is running: %v", err)
	}
	log.Info("IPFS is running")
	return nil
}

func checkGatewayFlag(do *definitions.Do) error {
	if do.Gateway != "" {
		_, err := url.Parse(do.Gateway)
		if err != nil {
			return fmt.Errorf("Invalid gateway URL provided: %v", err)
		}
		log.WithField("gateway", do.Gateway).Debug("Posting to")
	} else {
		log.Debug("Posting to gateway.ipfs.io")
	}
	return nil
}

// checks an ipfs hash to see if it is an object or a file
// returns true if an object (to be saved as a directory)
func isHashAnObject(ipfsHash string) (bool, error) {
	dirBool := false

	result, err := listFile(ipfsHash)
	if err != nil {
		return dirBool, err
	}
	if strings.TrimSpace(result) != "" { //not a dir
		return true, nil
	}

	return dirBool, nil
}
