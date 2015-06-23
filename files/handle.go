package files

import (
  "bytes"
  "fmt"
  "io"
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
  err := GetFilesRaw(args[0], args[1], cmd.Flags().Lookup("verbose").Changed, os.Stdout)
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
  err := PutFilesRaw(args[0], cmd.Flags().Lookup("verbose").Changed, os.Stdout)
  if err != nil {
    fmt.Println(err)
    os.Exit(1)
  }
}

func GetFilesRaw(hash, fileName string, verbose bool, w io.Writer) error {
  ipfsService := services.LoadServiceDefinition("ipfs")

  if services.IsServiceRunning(ipfsService.Service) {
    if verbose {
      w.Write([]byte("IPFS is running. Adding now."))
    }

    err := importFile(hash, fileName, verbose, w)
    if err != nil {
      return err
    }

  } else {
    if verbose {
      w.Write([]byte("IPFS is not running. Starting now."))
    }
    services.StartServiceByService(ipfsService.Service, ipfsService.Operations, verbose)

    err := importFile(hash, fileName, verbose, w)
    if err != nil {
      return err
    }
  }
  return nil
}

func PutFilesRaw(fileName string, verbose bool, w io.Writer) error {
  ipfsService := services.LoadServiceDefinition("ipfs")

  if services.IsServiceRunning(ipfsService.Service) {
    if verbose {
      w.Write([]byte("IPFS is running. Adding now."))
    }

    err := exportFile(fileName, verbose, w)
    if err != nil {
      return err
    }

  } else {
    if verbose {
      w.Write([]byte("IPFS is not running. Starting now."))
    }
    services.StartServiceByService(ipfsService.Service, ipfsService.Operations, verbose)

    err := exportFile(fileName, verbose, w)
    if err != nil {
      return err
    }
  }
  return nil
}

func importFile(hash, fileName string, verbose bool, w io.Writer) error {
  var err error
  if verbose {
    err = util.GetFromIPFS(hash, fileName, w)
  } else {
    err = util.GetFromIPFS(hash, fileName, bytes.NewBuffer([]byte{}))
  }

  if err != nil {
    return err
  }
  return nil
}

func exportFile(fileName string, verbose bool, w io.Writer) error {
  var hash string
  var err error

  if verbose {
    hash, err = util.SendToIPFS(fileName, w)
  } else {
    hash, err = util.SendToIPFS(fileName, bytes.NewBuffer([]byte{}))
  }
  if err != nil {
    return err
  }

  w.Write([]byte(hash + "\n"))
  return nil
}
