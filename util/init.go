package util


import (
  "fmt"
  "os"
  "os/exec"
  "path/filepath"

  "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common"
)

func Initialize(toPull, verbose bool) {
  if _, err := os.Stat(common.ErisRoot); err != nil {
    err := common.InitErisDir()
    if err != nil {
      fmt.Println(err)
      os.Exit(1)
    }
  } else {
    fmt.Printf("Root eris directory (%s) already exists. Please type `eris` to see the help.\n", common.ErisRoot)
    os.Exit(0)
  }

  if toPull {
    if err := pullRepo("eris-services", common.ServicesPath, verbose); err != nil {
      fmt.Println("Will just use only default IPFS def.")
      if e3 := ipfsDef(); e3 != nil {
        fmt.Println(err)
        os.Exit(1)
      }
      return
    } else {
      if err2 := pullRepo("eris-actions", common.ActionsPath, verbose); err2 != nil {
        fmt.Println(err)
        os.Exit(1)
      }
    }
  } else {
    if err := ipfsDef(); err != nil {
      fmt.Println(err)
      os.Exit(1)
    }
  }

  if verbose {
    fmt.Printf("Initialized eris root directory (%s) with default actions and service files\n", common.ErisRoot)
  }
}

func ipfsDef() error {
  if err := os.MkdirAll(common.ServicesPath, 0777); err != nil {
    return err
  }
  writer, err := os.Create(filepath.Join(common.ServicesPath, "ipfs.toml"))
  defer writer.Close()
  if err != nil {
    return err
  }
  ipfsD := `
[service]
name = "ipfs"
image = "eris/ipfs"
data_container = true
ports = ["4001:4001", "5001", "8080:8080"]

[maintainer]
name = "Eris Industries"
email = "support@erisindustries.com"

[location]
repository = "github.com/eris-ltd/eris-services"

[machine]
include = ["docker"]
requires = [""]
`
  writer.Write([]byte(ipfsD))
  return nil
}

func pullRepo(name, location string, verbose bool) error {
  src := "https://github.com/eris-ltd/" + name
  c := exec.Command("git", "clone", src, location)
  if verbose {
    c.Stdout = os.Stdout
    c.Stderr = os.Stderr
  }
  if err := c.Run(); err != nil {
    return err
  }
  return nil
}