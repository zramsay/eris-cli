package list

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"text/template"

	"github.com/monax/monax/definitions"
	"github.com/monax/monax/util"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/monax/monax/log"
)

const (
	// `monax ls` format.
	standardTmplHeader = "{{toupper .}}\tON\tVERSION"
	standardTmpl       = "{{.ShortName}}\t{{astmonaxk .Info.State.Running}}\t{{img .Info.Config.Image}}"

	// `monax ls -a` format.
	extendedTmplHeader = "{{toupper .}}\tON\tCONTAINER ID\tDATA CONTAINER\tIMAGE\tCOMMAND\tPORTS"
	extendedTmpl       = "{{.ShortName}}\t{{astmonaxk .Info.State.Running}}\t{{short .Info.ID}}\t{{short (dependent .ShortName)}}\t{{.Info.Config.Image}}\t{{.Info.Config.Cmd}}\t{{ports .Info}}"
)

var (
	monaxContainers = []*util.Details{}

	// Template helpers to manipulate raw field values in the output.
	helpers = map[string]interface{}{
		"toupper": func(word string) string {
			return strings.ToUpper(word)
		},
		"quote": func(word string) string {
			return strconv.Quote(word)
		},
		// Show a '*' symbol if a container is running.
		"astmonaxk": func(running bool) string {
			if running {
				return "*"
			}
			return "-"
		},
		// Truncate the longer ID version (handy for copying and pasting).
		"short": func(id string) string {
			if len(id) <= 10 {
				return id
			}

			return id[:10]
		},
		// Show a dependent data container name if it exists
		// for the given short name of a service or a chain.
		"dependent": func(name string) string {
			for _, container := range monaxContainers {
				if container.ShortName == name && container.Type == definitions.TypeData {
					return container.Info.ID
				}
			}
			return ""
		},
		// Pretty-format Docker ports.
		"ports": func(container *docker.Container) string {
			return util.FormulatePortsOutput(container)
		},
		"img": func(image string) string {
			tag := strings.Split(image, ":")
			if len(tag) == 2 {
				return tag[1]
			} else if len(tag) == 1 {
				return "latest"
			} else {
				return "unknown"
			}
		},
	}
)

// Containers display container information on the console in a format
// specified by the "format" parameter: the default "" and "extended" use the
// predefined Go templates, "json" dumps the JSON document of container
// details for every container. A custom format can be specified using
// the Go template syntax.
func Containers(t, format string, running bool) error {
	log.WithFields(log.Fields{
		"format": format,
		"type":   t,
	}).Debug("Listing containers")

	// Dump a JSON document then terminate.
	if format == "json" {
		return jsonContainers(t, running)
	}

	// Collect container information.
	util.MonaxContainers(func(name string, details *util.Details) bool {
		if running && !details.Info.State.Running && details.Type != definitions.TypeData {
			return false
		}
		monaxContainers = append(monaxContainers, details)
		return true
	}, false)

	// Keys for the parameter map.
	const (
		Standard = iota
		Extended
		Custom
	)
	key := Standard
	switch {
	case format == "extended":
		key = Extended
	case format != "":
		key = Custom
	}

	// Use a table to select template rendering parameters to avoid multiple nested ifs.
	buf := new(bytes.Buffer)
	renderParams := map[string]map[int][]struct {
		Type     string
		Header   string
		Template string
	}{
		definitions.TypeService: {
			Standard: {{t, standardTmplHeader, standardTmpl}},
			Extended: {{t, extendedTmplHeader, extendedTmpl}},
			Custom:   {{t, "", format}},
		},
		definitions.TypeChain: {
			Standard: {{t, standardTmplHeader, standardTmpl}},
			Extended: {{t, extendedTmplHeader, extendedTmpl}},
			Custom:   {{t, "", format}},
		},
		"all": {
			Standard: {
				{definitions.TypeService, standardTmplHeader, standardTmpl},
				{definitions.TypeChain, standardTmplHeader, standardTmpl},
			},
			Extended: {
				{definitions.TypeService, extendedTmplHeader, extendedTmpl},
				{definitions.TypeChain, extendedTmplHeader, extendedTmpl},
			},
			Custom: {
				{definitions.TypeService, "", format},
				{definitions.TypeChain, "", format},
			},
		},
	}

	if _, ok := renderParams[t]; !ok {
		return fmt.Errorf("Don't know the type %q to list containers for", t)
	}

	for _, p := range renderParams[t][key] {
		if err := render(buf, p.Type, p.Header, p.Template); err != nil {
			return err
		}
	}

	// 6 - minwidth, 1 - tabwidth (tab characters width), 5 - padding, ' ' - padchar, 0 - flags.
	tw := tabwriter.NewWriter(os.Stdout, 6, 1, 5, ' ', 0)
	buf.WriteTo(tw)
	tw.Flush()

	return nil
}

func render(buf *bytes.Buffer, t string, header, format string) error {
	r := strings.NewReplacer(`\t`, "\t", `\n`, "\n")
	if header != "" {
		tmplHeader, err := template.New("header").Funcs(helpers).Parse(r.Replace(header))
		if err != nil {
			return fmt.Errorf("Header template error: %v", err)
		}
		if err := tmplHeader.Execute(buf, t); err != nil {
			return fmt.Errorf("Header template exec error: %v", err)
		}
		buf.WriteString("\n")
	}

	tmplTable, err := template.New("containers").Funcs(helpers).Parse(r.Replace(format))
	if err != nil {
		return fmt.Errorf("Listing template error: %v", err)
	}

	for _, container := range monaxContainers {
		// Show containers for this section type.
		if container.Type != t {
			continue
		}

		if err := tmplTable.Execute(buf, container); err != nil {
			return fmt.Errorf("Listing template exec error: %v\n", err)
		}

		buf.WriteString("\n")
	}

	if header != "" {
		// Tabs are necessary so that the Tabwriter doesn't break
		// on a newline (1 tab per column).
		buf.WriteString("\t\t\t\t\t\t\n")
	}
	return nil
}

func jsonContainers(t string, running bool) error {
	// Collect container information.
	util.MonaxContainers(func(name string, details *util.Details) bool {
		if t == "all" || t == details.Type {
			monaxContainers = append(monaxContainers, details)
		}
		return true
	}, running)

	b, err := json.Marshal(monaxContainers)
	if err != nil {
		return err
	}
	var out bytes.Buffer
	json.Indent(&out, b, "", "  ")
	out.WriteTo(os.Stdout)
	io.WriteString(os.Stdout, "\n")

	return nil
}
