package files

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/services"

	log "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/ipfs"
)

func GetFiles(do *definitions.Do) error {
	ensureRunning()
	var err error
	if do.CSV != "" {
		log.WithFields(log.Fields{
			"from": do.CSV,
			"to":   do.NewName,
		}).Debug("Importing files")
		err = importFiles(do.CSV, do.NewName)

	} else {
		log.WithFields(log.Fields{
			"file": do.Name,
			"path": do.Path,
		}).Debug("Importing a file")
		err = importFile(do.Name, do.Path)
	}
	if err != nil {
		return err
	}
	do.Result = "success"
	return nil
}

func PutFiles(do *definitions.Do) error {
	ensureRunning()

	if do.Gateway != "" {
		_, err := url.Parse(do.Gateway)
		if err != nil {
			return fmt.Errorf("Invalid gateway URL provided %v\n", err)
		}
		log.WithField("gateway", do.Gateway).Debug("Posting to")
	} else {
		log.Debug("Posting to gateway.ipfs.io")
	}

	if do.AddDir {
		log.WithFields(log.Fields{
			"dir":     do.Name,
			"gateway": do.Gateway,
		}).Debug("Adding contents of a directory")
		hashes, err := exportDir(do.Name, do.Gateway)
		if err != nil {
			return err
		}
		do.Result = hashes
	} else {
		log.WithFields(log.Fields{
			"file":    do.Name,
			"gateway": do.Gateway,
		}).Debug("Adding a file")
		hash, err := exportFile(do.Name, do.Gateway)
		if err != nil {
			return err
		}
		do.Result = hash
	}
	return nil
}

func PinFiles(do *definitions.Do) error {
	ensureRunning()
	if do.CSV != "" {
		log.WithField("=>", do.CSV).Debug("Pinning all files from")
		hashes, err := pinFiles(do.CSV)
		if err != nil {
			return err
		}
		do.Result = hashes

	} else {
		log.WithFields(log.Fields{
			"file": do.Name,
			"path": do.Path,
		}).Debug("Pinning a file")
		hash, err := pinFile(do.Name)
		if err != nil {
			return err
		}
		do.Result = hash
	}
	return nil
}

func CatFiles(do *definitions.Do) error {
	ensureRunning()
	log.WithFields(log.Fields{
		"file": do.Name,
		"path": do.Path,
	}).Debug("Dumping the contents of a file")
	hash, err := catFile(do.Name)
	if err != nil {
		return err
	}
	do.Result = hash
	return nil
}

func ListFiles(do *definitions.Do) error {
	ensureRunning()
	log.WithFields(log.Fields{
		"file": do.Name,
		"path": do.Path,
	}).Debug("Listing an object")
	hash, err := listFile(do.Name)
	if err != nil {
		return err
	}
	do.Result = hash
	return nil
}

func ManagePinned(do *definitions.Do) error {
	ensureRunning()
	if do.Rm && do.Hash != "" {
		return fmt.Errorf("Either remove a file by hash or all of them\n")
	}

	if do.Rm {
		log.Info("Removing all cached files")
		hashes, err := rmAllPinned()
		if err != nil {
			return err
		}
		do.Result = hashes
	} else if do.Hash != "" {
		log.WithField("hash", do.Hash).Info("Removing from cache")
		hashes, err := rmPinnedByHash(do.Hash)
		if err != nil {
			return err
		}
		do.Result = hashes
	} else {
		log.Debug("Listing files pinned locally")
		hash, err := listPinned()
		if err != nil {
			return err
		}
		do.Result = hash
	}
	return nil
}

func importFile(hash, fileName string) error {
	var err error

	if log.GetLevel() > 0 {
		err = ipfs.GetFromIPFS(hash, fileName, "", os.Stdout)
	} else {
		err = ipfs.GetFromIPFS(hash, fileName, "", bytes.NewBuffer([]byte{}))
	}
	if err != nil {
		return err
	}
	return nil
}

func importFiles(csvfile, newdir string) error {
	var err error

	csvFile, err := os.Open(csvfile)
	if err != nil {
		return fmt.Errorf("error opening csv file: %v\n", err)
	}
	defer csvFile.Close()

	reader := csv.NewReader(csvFile)
	rawCSVdata, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("error reading csv file: %v\n", err)
	}

	for _, each := range rawCSVdata {
		if log.GetLevel() > 0 {
			err = ipfs.GetFromIPFS(each[0], each[1], newdir, os.Stdout)
		} else {
			err = ipfs.GetFromIPFS(each[0], each[1], newdir, bytes.NewBuffer([]byte{}))
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func exportFile(fileName, gateway string) (string, error) {
	var hash string
	var err error

	if log.GetLevel() > 0 {
		hash, err = ipfs.SendToIPFS(fileName, gateway, os.Stdout)
	} else {
		hash, err = ipfs.SendToIPFS(fileName, gateway, bytes.NewBuffer([]byte{}))
	}
	if err != nil {
		return "", err
	}

	return hash, nil
}

func exportDir(dirName, gateway string) (string, error) {
	var hashes string
	var err error

	files, err := ioutil.ReadDir(dirName)
	if err != nil {
		return "", fmt.Errorf("error reading directory %v\n", err)
	}
	gwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("error getting working directory %v\n", err)
	}
	hashArray := make([]string, len(files))
	fileNames := make([]string, len(files))
	//the dir ends up in the loop & tries to post
	for i := range files {
		//hacky
		file := path.Join(gwd, dirName, files[i].Name())
		if log.GetLevel() > 0 {
			hashArray[i], err = ipfs.SendToIPFS(file, gateway, os.Stdout)
		} else {
			hashArray[i], err = ipfs.SendToIPFS(file, gateway, bytes.NewBuffer([]byte{}))
		}
		if err != nil {
			return "", fmt.Errorf("error reading file %v\n", err)
		}
		fileNames[i] = files[i].Name()
	}

	err = writeCsv(hashArray, fileNames)
	if err != nil {
		return "", err
	}

	hashes = strings.Join(hashArray, "\n")

	return hashes, nil
}

func pinFile(fileHash string) (string, error) {
	var hash string
	var err error

	if log.GetLevel() > 0 {
		hash, err = ipfs.PinToIPFS(fileHash, os.Stdout)
	} else {
		hash, err = ipfs.PinToIPFS(fileHash, bytes.NewBuffer([]byte{}))
	}
	if err != nil {
		return "", err
	}
	return hash, nil
}

func pinFiles(csvfile string) (string, error) {
	var err error

	csvFile, err := os.Open(csvfile)
	if err != nil {
		return "", fmt.Errorf("error opening csv file: %v\n", err)
	}
	defer csvFile.Close()

	reader := csv.NewReader(csvFile)
	rawCSVdata, err := reader.ReadAll()
	if err != nil {
		return "", fmt.Errorf("error reading csv file: %v\n", err)
	}

	hashArray := make([]string, len(rawCSVdata))
	for i, each := range rawCSVdata {
		if log.GetLevel() > 0 {
			hashArray[i], err = ipfs.PinToIPFS(each[0], os.Stdout)
		} else {
			hashArray[i], err = ipfs.PinToIPFS(each[0], bytes.NewBuffer([]byte{}))
		}
		if err != nil {
			return "", err
		}
	}
	hashes := strings.Join(hashArray, "\n")
	return hashes, nil
}

func catFile(fileHash string) (string, error) {
	var hash string
	var err error
	if log.GetLevel() > 0 {
		hash, err = ipfs.CatFromIPFS(fileHash, os.Stdout)
	} else {
		hash, err = ipfs.CatFromIPFS(fileHash, bytes.NewBuffer([]byte{}))
	}
	if err != nil {
		return "", err
	}
	return hash, nil
}

func listFile(objectHash string) (string, error) {
	var hash string
	var err error
	if log.GetLevel() > 0 {
		hash, err = ipfs.ListFromIPFS(objectHash, os.Stdout)
	} else {
		hash, err = ipfs.ListFromIPFS(objectHash, bytes.NewBuffer([]byte{}))
	}
	if err != nil {
		return "", err
	}
	return hash, nil
}

func listPinned() (string, error) {
	var hash string
	var err error
	if log.GetLevel() > 0 {
		hash, err = ipfs.ListPinnedFromIPFS(os.Stdout)
	} else {
		hash, err = ipfs.ListPinnedFromIPFS(bytes.NewBuffer([]byte{}))
	}
	if err != nil {
		return "", err
	}
	return hash, nil
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
	var err error
	if log.GetLevel() > 0 {
		hash, err = ipfs.RemovePinnedFromIPFS(hash, os.Stdout)
	} else {
		hash, err = ipfs.RemovePinnedFromIPFS(hash, bytes.NewBuffer([]byte{}))
	}
	if err != nil {
		return "", err
	}
	return hash, nil
}

//---------------------------------------------------------
// helpers

func writeCsv(hashArray, fileNames []string) error {
	strToWrite := make([][]string, len(hashArray))
	for i := range hashArray {
		strToWrite[i] = []string{hashArray[i], fileNames[i]}

	}

	csvfile, err := os.Create("ipfs_hashes.csv")
	if err != nil {
		return fmt.Errorf("error creating csv file:", err)
	}
	defer csvfile.Close()

	w := csv.NewWriter(csvfile)
	w.WriteAll(strToWrite)

	if err := w.Error(); err != nil {
		return fmt.Errorf("error writing csv: \n", err)
	}
	return nil
}

func ensureRunning() {
	doNow := definitions.NowDo()
	doNow.Name = "ipfs"
	err := services.EnsureRunning(doNow)
	if err != nil {
		fmt.Printf("Failed to ensure IPFS is running: %v", err)
		return
	}
	log.Info("IPFS is running")
}
