package files

import (
	"bytes"
	"fmt"
	"os"

	"github.com/eris-ltd/eris-cli/services"
	"github.com/eris-ltd/eris-cli/util"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/cobra"
)

func Get(cmd *cobra.Command, args []string) {
	if len(args) != 2 {
		fmt.Println("Please give me: eris files get ipfsHASH fileName")
		os.Exit(1)
	}
	err := GetFilesRaw(args[0], args[1])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func Put(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		fmt.Println("Please give me: eris files put fileName")
		os.Exit(1)
	}
	err := PutFilesRaw(args[0])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func GetFilesRaw(hash, fileName string) error {
	ipfsService, err := services.LoadServiceDefinition("ipfs")
	if err != nil {
		return err
	}

	if services.IsServiceRunning(ipfsService.Service) {
		logger.Infoln("IPFS is running. Adding now.")

		err := importFile(hash, fileName)
		if err != nil {
			return err
		}

	} else {
		logger.Infoln("IPFS is not running. Starting now.")
		err := services.StartServiceByService(ipfsService.Service, ipfsService.Operations)
		if err != nil {
			return err
		}

		err = importFile(hash, fileName)
		if err != nil {
			return err
		}
	}
	return nil
}

func PutFilesRaw(fileName string) error {
	ipfsService, err := services.LoadServiceDefinition("ipfs")
	if err != nil {
		return err
	}

	if services.IsServiceRunning(ipfsService.Service) {
		logger.Infoln("IPFS is running. Adding now.")

		err := exportFile(fileName)
		if err != nil {
			return err
		}

	} else {
		logger.Infoln("IPFS is not running. Starting now.")
		if err := services.StartServiceByService(ipfsService.Service, ipfsService.Operations); err != nil {
			return err
		}

		if err := exportFile(fileName); err != nil {
			return err
		}
	}
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

func exportFile(fileName string) error {
	var hash string
	var err error

	if logger.Level > 0 {
		hash, err = util.SendToIPFS(fileName, logger.Writer)
	} else {
		hash, err = util.SendToIPFS(fileName, bytes.NewBuffer([]byte{}))
	}
	if err != nil {
		return err
	}

	logger.Println(hash)
	return nil
}
