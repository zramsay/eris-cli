package main

// TODO significant update!
import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/eris-ltd/common/go/common"
	"github.com/eris-ltd/common/go/docs"
	commands "github.com/eris-ltd/eris-cm/cmd"
	"github.com/eris-ltd/eris-cm/version"

	"github.com/BurntSushi/toml"
	"github.com/spf13/cobra"
)

var RENDER_DIR = fmt.Sprintf("./docs/eris-cm/%s/", version.VERSION)

var SPECS_DIR = "./docs/"

var BASE_URL = fmt.Sprintf("https://monax.io/docs/documentation/cm/%s/", version.VERSION)

const FRONT_MATTER = `---

layout:     documentation
title:      "Documentation | eris:chain_manager | {{}}"

---

`

const ACCOUNT_TYPES_INTRO = `In order to reduce the complexity of dealing with permissioning
of chains, eris chains uses the concept of account_types. Account Types are
simple default bundles of permissions, tokens, and names which are used
as an abstraction when building eris chains so as to reduce the complexity
of that process.

This document contains the default account types which are used by
eris chains make when creating the necessary files for a new chain. Users
have an ability to add additional account_types which will be needed for
their purposes in a very easy manner.

The defaults provided by eris:chain_manager should be thought of as simply
that, defaults, rather than as restrictive ("we only get these") manner.

See also [chain_types](chain_types) specification.
`

const CHAIN_TYPES_INTRO = `In order to reduce the complexity of dealing with permissioning
of chains, eris chains uses the concept of chain_types. Chain Types are
bundles of [account_types](account_types). They define the number of
each account type which is required to make the given chain_type.

In the future as we continue to add more optionality to eris chains at
the consensus engine and application levels of the eris chain more
functionality will be added to chain types.
`

type AccountT struct {
	Name        string `mapstructure:"name" json:"name" yaml:"name" toml:"name"`
	Definition  string `mapstructure:"definition" json:"definition" yaml:"definition" toml:"definition"`
	TypicalUser string `mapstructure:"typical_user" json:"typical_user" yaml:"typical_user" toml:"typical_user"`
}

type ChainT struct {
	Name         string         `mapstructure:"name" json:"name" yaml:"name" toml:"name"`
	Definition   string         `mapstructure:"definition" json:"definition" yaml:"definition" toml:"definition"`
	AccountTypes map[string]int `mapstructure:"account_types" json:"account_types" yaml:"account_types" toml:"account_types"`
}

// Repository maintainers should customize the next two lines.
var Description = "Chain Manager Tooling"                                   // should match the docs site name
var RenderDir = fmt.Sprintf("./docs/documentation/cm/%s/", version.VERSION) // should be the "shortversion..."

// The below variables should be updated only if necessary.
var Specs = []*docs.Entry{}
var Examples = []*docs.Entry{}
var SpecsDir = "./docs/specs"
var ExamplesDir = "./docs/examples"

type Cmd struct {
	Command     *cobra.Command
	Entry       *docs.Entry
	Description string
}

func generateAccountTypes() {
	generatedFile := filepath.Join("docs", "specs", "account_types.md")
	accountDescriptions := []string{"# Default eris chains Account Types", ACCOUNT_TYPES_INTRO}
	accountTypeFiles, _ := filepath.Glob(filepath.Join(common.ErisGo, "eris-cm", "account_types", "*"))
	for _, file := range accountTypeFiles {
		var thisAct AccountT
		tomlData, _ := ioutil.ReadFile(file)
		_, _ = toml.Decode(string(tomlData), &thisAct)
		accountDescriptions = append(accountDescriptions, fmt.Sprintf("# The %s Account Type", thisAct.Name))
		accountDescriptions = append(accountDescriptions, thisAct.Definition)
		accountDescriptions = append(accountDescriptions, fmt.Sprintf("## Typical Users of this Account Type: %s", thisAct.Name))
		accountDescriptions = append(accountDescriptions, thisAct.TypicalUser)
	}
	ioutil.WriteFile(generatedFile, []byte(strings.Join(accountDescriptions, "\n\n")), 0644)
}

func generateChainTypes() {
	generatedFile := filepath.Join("docs", "specs", "chain_types.md")
	chainDescriptions := []string{"# Default eris chains Chain Types", CHAIN_TYPES_INTRO}
	chainTypeFiles, _ := filepath.Glob(filepath.Join(common.ErisGo, "eris-cm", "chain_types", "*"))
	for _, file := range chainTypeFiles {
		var thisChain ChainT
		tomlData, _ := ioutil.ReadFile(file)
		_, _ = toml.Decode(string(tomlData), &thisChain)
		chainDescriptions = append(chainDescriptions, fmt.Sprintf("# The %s Chain Type", thisChain.Name))
		chainDescriptions = append(chainDescriptions, thisChain.Definition)
		chainDescriptions = append(chainDescriptions, fmt.Sprintf("## Number of Account Types for Chain Type: %s", thisChain.Name))
		numbers := ""
		for name, number := range thisChain.AccountTypes {
			numbers = fmt.Sprintf("%s\n%s:%03d", numbers, name, number)
		}
		chainDescriptions = append(chainDescriptions, numbers)
	}
	ioutil.WriteFile(generatedFile, []byte(strings.Join(chainDescriptions, "\n\n")), 0644)
}

func RenderFiles(cmdRaw *cobra.Command, tmpl *template.Template) error {
	this_entry := &docs.Entry{
		Title:          cmdRaw.CommandPath(),
		Specifications: Specs,
		Examples:       Examples,
		BaseURL:        strings.Replace(RenderDir, ".", "", 1),
		Template:       tmpl,
		FileName:       docs.GenerateFileName(RenderDir, cmdRaw.CommandPath()),
	}

	cmd := &Cmd{
		Command:     cmdRaw,
		Entry:       this_entry,
		Description: Description,
	}

	for _, command := range cmd.Command.Commands() {
		RenderFiles(command, tmpl)
	}

	if !cmd.Command.HasParent() {
		entries := append(cmd.Entry.Specifications, cmd.Entry.Examples...)
		for _, entry := range entries {
			entry.Specifications = cmd.Entry.Specifications
			entry.Examples = cmd.Entry.Examples
			entry.CmdEntryPoint = cmd.Entry.Title
			entry.BaseURL = cmd.Entry.BaseURL
			if err := docs.RenderEntry(entry); err != nil {
				return err
			}
		}
	}

	outFile, err := os.Create(cmd.Entry.FileName)
	if err != nil {
		return err
	}
	defer outFile.Close()

	err = cmd.Entry.Template.Execute(outFile, cmd)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	// Repository maintainers should populate the top level command object.
	cm := commands.ErisCMCmd
	commands.InitErisChainManager()
	commands.AddGlobalFlags()
	commands.AddCommands()

	// Make the proper directory.
	var err error
	if _, err = os.Stat(RenderDir); os.IsNotExist(err) {
		err = os.MkdirAll(RenderDir, 0755)
		if err != nil {
			panic(err)
		}
	}

	generateAccountTypes()
	generateChainTypes()

	// Generate specs and examples files.
	Specs, err = docs.GenerateEntries(SpecsDir, (RenderDir + "specifications/"), Description)
	if err != nil {
		panic(err)
	}
	Examples, err = docs.GenerateEntries(ExamplesDir, (RenderDir + "examples/"), Description)
	if err != nil {
		panic(err)
	}

	// Get template from docs generator.
	tmpl, err := docs.GenerateCommandsTemplate()
	if err != nil {
		panic(err)
	}

	// Render the templates.
	if err = RenderFiles(cm, tmpl); err != nil {
		panic(err)
	}
}
