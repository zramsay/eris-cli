package util

import (
	"fmt"
	"sort"
	"strings"
	"text/template"

	"github.com/monax/monax/config"
	"github.com/monax/monax/log"

	docker "github.com/fsouza/go-dockerclient"

	"github.com/oleiade/reflections"
	"github.com/serenize/snaker"
)

func PrintInspectionReport(cont *docker.Container, field string) error {
	switch field {
	case "line":
		parts, err := printLine(cont, false)
		if err != nil {
			return err
		}
		log.Warn(strings.Join(parts, " "))
	case "all":
		for _, obj := range []interface{}{cont, cont.Config, cont.HostConfig, cont.NetworkSettings} {
			t, err := reflections.Fields(obj)
			if err != nil {
				return fmt.Errorf("The marmots had an error trying to print a nice report\n%s", err)
			}
			for _, f := range t {
				printReport(obj, f)
			}
		}
	default:
		return printField(cont, field)
	}

	return nil
}

func PrintLineByContainerID(containerID string, existing bool) ([]string, error) {
	cont, err := DockerClient.InspectContainer(containerID)
	if err != nil {
		return nil, DockerError(err)
	}
	return printLine(cont, existing)
}

// this function populates the listing functions only for flags/tests
func printLine(container *docker.Container, existing bool) ([]string, error) {
	tmp, err := reflections.GetField(container, "Name")
	if err != nil {
		return nil, err
	}
	n := tmp.(string)

	var running string
	if !existing {
		running = "Yes"
	} else {
		running = "No"
	}

	details := ContainerDetails(n)

	parts := []string{details.ShortName, "", running, details.FullName, FormulatePortsOutput(container)}

	return parts, nil
}

// this function is for parsing single variables
func printField(container interface{}, field string) error {
	log.WithField("=>", field).Debug("Inspecting field")
	var line string

	// We allow fields to be passed using dot syntax, but
	// we have to make sure all fields are Camelized
	lineSplit := strings.Split(field, ".")
	for n, f := range lineSplit {
		lineSplit[n] = camelize(f)
	}
	FieldCamel := strings.Join(lineSplit, ".")

	f, _ := reflections.GetFieldKind(container, FieldCamel)
	log.Debug("Field type", f)
	switch f.String() {
	case "ptr":
		//we don't recurse into to gain a bit more control... this function will be rarely used and doesn't have to be perfectly parseable.
	case "map":
		line = fmt.Sprintf("{{ range $key, $val := .%v }}{{ $key }}->{{ $val }}\n{{ end }}\n", FieldCamel)
	case "slice":
		line = fmt.Sprintf("{{ range .%v }}{{ . }}\n{{ end }}\n", FieldCamel)
	default:
		line = fmt.Sprintf("{{.%v}}\n", FieldCamel)
	}
	return writeTemplate(container, line)
}

// this function is more verbose and used when inspect is
// set to all
func printReport(container interface{}, field string) error {
	var line string
	FieldCamel := camelize(field)
	f, _ := reflections.GetFieldKind(container, FieldCamel)
	switch f.String() {
	case "ptr":
		// we don't recurse into to gain a bit more control... this function will be rarely used and doesn't have to be perfectly parseable.
	case "map":
		line = fmt.Sprintf("%-20s\n{{ range $key, $val := .%v }}%20v{{ $key }}->{{ $val }}\n{{ end }}", FieldCamel+":", FieldCamel, "")
	case "slice":
		line = fmt.Sprintf("%-20s\n{{ range .%v }}%20v{{ . }}\n{{ end }}", FieldCamel+":", FieldCamel, "")
	default:
		line = fmt.Sprintf("%-20s{{.%v}}\n", FieldCamel+":", FieldCamel)
	}
	return writeTemplate(container, line)
}

func FormulatePortsOutput(container *docker.Container) string {
	var ports string
	for k, v := range container.NetworkSettings.Ports {
		if len(v) != 0 {
			ports = ports + fmt.Sprintf("%v:%v->%v ", v[0].HostIP, v[0].HostPort, k) // published ports
		} else {
			ports = ports + fmt.Sprintf("%v ", k)
		}
	}

	split := strings.Split(strings.Trim(ports, " "), " ")
	sort.Sort(sort.StringSlice(split))

	return strings.Join(split, ", ")
}

func camelize(field string) string {
	return snaker.SnakeToCamel(field)
}

func writeTemplate(container interface{}, toParse string) error {
	log.WithField("=>", strings.Replace(toParse, "\n", " ", -1)).Info("Template parsing")
	tmpl, err := template.New("field").Parse(toParse)
	if err != nil {
		return err
	}

	if err = tmpl.Execute(config.Global.Writer, container); err != nil {
		return err
	}

	return nil
}
