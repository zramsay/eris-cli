package files

import (
	"bytes"
	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/loaders"
	"github.com/eris-ltd/eris-cli/perform"
	"github.com/eris-ltd/eris-cli/services"
	"github.com/eris-ltd/eris-cli/util"
	"time"
)

func GetFiles(do *definitions.Do) error {
	ipfsService, err := loaders.LoadServiceDefinition("ipfs", false, 1)
	if err != nil {
		return err
	}

	if services.IsServiceRunning(ipfsService.Service, ipfsService.Operations) {
		logger.Infoln("IPFS is running. Adding now.")

		logger.Debugf("Gonna Import a file =>\t\t%s:%v\n", do.Name, do.Path)
		err := importFile(do.Name, do.Path)
		if err != nil {
			return err
		}

	} else {
		logger.Infoln("IPFS is not running. Starting now. Waiting for IPFS to become available")
		time.Sleep(time.Second * 5) // this is dirty
		err := perform.DockerRun(ipfsService.Service, ipfsService.Operations)
		if err != nil {
			return err
		}

		err = importFile(do.Args[0], do.Name)
		if err != nil {
			return err
		}
	}
	do.Result = "success"
	return nil
}

func PutFiles(do *definitions.Do) error {
	var hash string
	var err error

	ipfsService, err := loaders.LoadServiceDefinition("ipfs", false, 1)
	if err != nil {
		return err
	}

	if services.IsServiceRunning(ipfsService.Service, ipfsService.Operations) {
		logger.Infoln("IPFS is running. Adding now.")

		hash, err = exportFile(do.Name)
		if err != nil {
			return err
		}

	} else {
		logger.Infoln("IPFS is not running. Starting now.")
		time.Sleep(time.Second * 5) // this is dirty

		if err := perform.DockerRun(ipfsService.Service, ipfsService.Operations); err != nil {
			return err
		}

		hash, err = exportFile(do.Name)
		if err != nil {
			return err
		}
	}
	do.Result = hash
	return nil
}

func PinFiles(do *definitions.Do) error {
	var hash string
	var err error

	ipfsService, err := loaders.LoadServiceDefinition("ipfs", false, 1)
	if err != nil {
		return err
	}

	if services.IsServiceRunning(ipfsService.Service, ipfsService.Operations) {
		logger.Infoln("IPFS is running. Pining now.")
		hash, err = pinFile(do.Name)
		if err != nil {
			return err
		}

	} else {
		logger.Infoln("IPFS is not running. Starting now.")
		time.Sleep(time.Second * 5) // this is dirty

		if err := perform.DockerRun(ipfsService.Service, ipfsService.Operations); err != nil {
			return err
		}

		hash, err = pinFile(do.Name)
		if err != nil {
			return err
		}
	}
	do.Result = hash
	return nil
}

func importFile(hash, fileName string) error {
	var err error
	if logger.Level > 0 {
		err = util.GetFromIPFS(hash, fileName, logger.Writer)
	} else {
		err = util.GetFromIPFS(hash, fileName, bytes.NewBuffer([]byte{}))
	}

	if err != nil {
		return err
	}
	return nil
}

func exportFile(fileName string) (string, error) {
	var hash string
	var err error

	if logger.Level > 0 {
		hash, err = util.SendToIPFS(fileName, logger.Writer)
	} else {
		hash, err = util.SendToIPFS(fileName, bytes.NewBuffer([]byte{}))
	}
	if err != nil {
		return "", err
	}

	return hash, nil
}

func pinFile(fileHash string) (string, error) {
	var hash string
	var err error

	if logger.Level > 0 {
		hash, err = util.PinToIPFS(fileHash, logger.Writer)
	} else {
		hash, err = util.PinToIPFS(fileHash, bytes.NewBuffer([]byte{}))
	}
	if err != nil {
		return "", err
	}

	return hash, nil

}
