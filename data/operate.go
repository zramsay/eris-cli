package data

import (
  "fmt"
  "io"
  "os"
  "path/filepath"

  "github.com/eris-ltd/eris-cli/perform"
  "github.com/eris-ltd/eris-cli/util"

  "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/ebuchman/go-shell-pipes"
  dir "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common"

  "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"
  "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/cobra"
)

func Import(cmd *cobra.Command, args []string) {
  checkServiceGiven(args)
  ImportDataRaw(args[0], cmd.Flags().Lookup("verbose").Changed)
}

func Export(cmd *cobra.Command, args []string) {
  checkServiceGiven(args)
  ExportDataRaw(args[0], cmd.Flags().Lookup("verbose").Changed)
}

func ImportDataRaw(name string, verbose bool) {
  if parseKnown(name) {

    containerName := nameToContainerName(name)
    importPath := filepath.Join(dir.DataContainersPath, name)

    // temp until docker cp works both ways.
    os.Chdir(importPath)
    cmd := "tar cf - . | docker run -i --rm --volumes-from " + containerName + " eris/data tar xf - -C /home/eris/.eris"

    s, err := pipes.RunString(cmd)
    if err != nil {
      fmt.Println(err)
      os.Exit(1)
    }

    if verbose {
      fmt.Println(s)
    }
  } else {
    fmt.Println("I cannot find that data container. Please check the data container name you sent me.")
  }
}

func ExportDataRaw(name string, verbose bool) {
  if parseKnown(name) {
    if verbose {
      fmt.Println("Exporting data container", name)
    }

    exportPath := filepath.Join(dir.DataContainersPath, name)
    _, ops := mockService(name)
    service, exists := perform.ContainerExists(ops)

    if !exists {
      fmt.Println("There is no data container for that service.")
      os.Exit(0)
    }
    if verbose {
      fmt.Println("Service ID: " + service.ID)
    }

    cont, err := util.DockerClient.InspectContainer(service.ID)
    if err != nil {
      // TODO: better error handling
      fmt.Printf("Failed to inspect container (%s) ->\n  %v\n", service.ID, err)
      os.Exit(1)
    }

    var dest string
    vol := cont.Volumes
    for k, v := range vol {
      if k == "/home/eris/.eris" {
        dest = filepath.Base(v)
      }
    }
    dest = filepath.Join(exportPath, dest)

    reader, writer := io.Pipe()
    opts := docker.CopyFromContainerOptions{
      OutputStream: writer,
      Container:    service.ID,
      Resource:     "/home/eris/.eris/",
    }

    go func() {
      err := util.DockerClient.CopyFromContainer(opts)
      if err != nil {
        fmt.Println(err)
        os.Exit(1)
      }
      writer.Close()
    }()
    _ = util.Untar(reader, name, exportPath)

    var toMove []string
    os.Chdir(exportPath)
    toMove, err = filepath.Glob(filepath.Join(dest, "*"))
    if err != nil {
      fmt.Println(err)
      os.Exit(1)
    }

    for _, f := range toMove {
      err =  os.Rename(f, filepath.Join(exportPath, filepath.Base(f)))
      if err != nil {
        // TODO: better errors
        fmt.Println(err)
        continue
      }
    }
    err = os.RemoveAll(dest)

  } else {
    if verbose {
      fmt.Println("I cannot find that data container. Please check the data container name you sent me.")
    }
  }
}