package list

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"text/template"

	"github.com/monax/monax/util"

	"github.com/docker/docker/pkg/term"
	"github.com/kr/text/colwriter"
)

// Definition holds useful data about definition files
// (the definition name and full path to the file).
type Definition struct {
	Name       string
	Definition string
}

// Known list definition files for a given type t ("services") from the Monax root directory in one of the 2 formats,
// specified by the format parameter. Default is `ls(1)` multicolumn format,
// `json` dumps the JSON document onto the console. A custom format can
// be specified using the `text/template` Go package syntax, e.g.:
//
//  `{{.Name}}`
//  `{{.Name}}\t{{.Definition}}`
//
func Known(t, format string) error {
	var definitions []Definition

	for _, file := range util.GetGlobalLevelConfigFilesByType(t, true) {
		definitions = append(definitions, Definition{
			Name:       strings.TrimSuffix(filepath.Base(file), filepath.Ext(file)),
			Definition: file,
		})
	}

	switch {
	case format == "json":
		return jsonKnown(definitions)
	case format != "":
		return customKnown(definitions, format)
	}

	return columnizeKnown(definitions)
}

func jsonKnown(definitions []Definition) error {
	b, err := json.Marshal(definitions)
	if err != nil {
		return err
	}
	var out bytes.Buffer
	json.Indent(&out, b, "", "  ")
	out.WriteTo(os.Stdout)
	io.WriteString(os.Stdout, "\n")

	return nil
}

func columnizeKnown(definitions []Definition) error {
	// Terminal column width, default is 80.
	columns := 80
	if winsize, err := term.GetWinsize(os.Stdout.Fd()); err == nil {
		columns = int(winsize.Width)
	}

	w := colwriter.NewWriter(os.Stdout, columns, 0)
	for _, definition := range definitions {
		// Extra space to match the standard `ls` multicolumn output
		// which uses 2 spaces between file names.
		fmt.Fprintln(w, filepath.Base(definition.Definition)+" ")
	}
	w.Flush()

	return nil
}

func customKnown(definitions []Definition, format string) error {
	r := strings.NewReplacer(`\t`, "\t", `\n`, "\n")
	tmpl, err := template.New("known").Parse(r.Replace(format))
	if err != nil {
		return fmt.Errorf("Template error: %v", err)
	}

	buf := new(bytes.Buffer)
	for _, definition := range definitions {
		if err := tmpl.Execute(buf, definition); err != nil {
			return fmt.Errorf("Template exec error: %v", err)
		}
		buf.WriteString("\n")
	}

	// 6 - minwidth, 1 - tabwidth (tab characters width), 5 - padding, ' ' - padchar, 0 - flags.
	tw := tabwriter.NewWriter(os.Stdout, 6, 1, 5, ' ', 0)
	buf.WriteTo(tw)
	tw.Flush()

	return nil
}
