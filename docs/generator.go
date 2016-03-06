package main

import (
	"fmt"

	commands "github.com/eris-ltd/eris-cli/cmd"
	"github.com/eris-ltd/eris-cli/version"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/common"
)

var RENDER_DIR = fmt.Sprintf("./docs/eris-cli/%s/", version.VERSION)

var SPECS_DIR = "./docs/"

var BASE_URL = fmt.Sprintf("https://docs.erisindustries.com/documentation/eris-cli/%s/", version.VERSION)

const FRONT_MATTER = `---

layout:     documentation
title:      "Documentation | eris:cli | {{}}"

---

`

func main() {
	eris := commands.ErisCmd
	commands.InitializeConfig()
	commands.AddGlobalFlags()
	commands.AddCommands()
	specs := common.GenerateSpecs(SPECS_DIR, RENDER_DIR, FRONT_MATTER)
	common.GenerateTree(eris, RENDER_DIR, specs, FRONT_MATTER, BASE_URL)
}
