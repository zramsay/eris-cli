package main

import (
  // "bytes"
  // "fmt"
  // "os"

  "github.com/eris-ltd/eris-cli/commands"

  "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/cobra"
)

func main() {
    eris := commands.ErisCmd
    commands.AddCommands()

    cobra.GenMarkdownTree(eris, "./rendered/")

    // out := new(bytes.Buffer)
    // cobra.GenMarkdown(eris, out)
    // _, _ = out.WriteString("\n\n")
    // cobra.GenMarkdown(commands.Projects, out)
    // _, _ = out.WriteString("\n\n")
    // cobra.GenMarkdown(commands.Chains, out)
    // _, _ = out.WriteString("\n\n")
    // cobra.GenMarkdown(commands.Services, out)
    // _, _ = out.WriteString("\n\n")
    // cobra.GenMarkdown(commands.Actions, out)
    // _, _ = out.WriteString("\n\n")
    // cobra.GenMarkdown(commands.Remotes, out)
    // _, _ = out.WriteString("\n\n")
    // cobra.GenMarkdown(commands.Keys, out)
    // _, _ = out.WriteString("\n\n")
    // cobra.GenMarkdown(commands.Config, out)
    // _, _ = out.WriteString("\n\n")
    // cobra.GenMarkdown(commands.Version, out)

    // outFile, err := os.Create("./eris.md")
    // if err != nil {
    //   fmt.Println(err)
    //   os.Exit(1)
    // }
    // defer outFile.Close()

    // _, err = outFile.Write(out.Bytes())
    // if err != nil {
    //   fmt.Println(err)
    //   os.Exit(1)
    // }
}