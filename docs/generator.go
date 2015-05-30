package main

import (
  "bytes"
  "fmt"
  // "io/ioutil"
  "os"

  "github.com/eris-ltd/eris-cli/commands"

  "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/cobra"
)

func main() {
    out := new(bytes.Buffer)
    eris := commands.ErisCmd
    commands.AddCommands()
    cobra.GenMarkdown(eris, out)
    cobra.GenMarkdown(commands.Projects, out)
    cobra.GenMarkdown(commands.Services, out)
    cobra.GenMarkdown(commands.Chains, out)
    // cobra.GenMarkdown(commands.Keys, out)
    // cobra.GenMarkdown(commands.Remotes, out)
    cobra.GenMarkdown(commands.Config, out)
    cobra.GenMarkdown(commands.Version, out)

    outFile, err := os.Create("./eris.md")
    if err != nil {
      fmt.Println(err)
      os.Exit(1)
    }
    defer outFile.Close()

    _, err = outFile.Write(out.Bytes())
    if err != nil {
      fmt.Println(err)
      os.Exit(1)
    }
}