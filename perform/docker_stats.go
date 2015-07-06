package perform

import (
	"fmt"
	"sort"
	"strings"
	"text/template"
	"unicode"

	"github.com/eris-ltd/eris-cli/util"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/oleiade/reflections"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/serenize/snaker"
)

func PrintInspectionReport(cont *docker.Container, field string) error {
	switch field {
	case "line":
		return PrintLineByContainerID(cont.ID)
	case "all":
		for _, obj := range []interface{}{cont, cont.Config, cont.HostConfig, cont.NetworkSettings} {
			t, err := reflections.Fields(obj)
			if err != nil {
				return fmt.Errorf("The PrintInspectionReport marmot had an error getting the fields using reflection.Fields\n%s", err)
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

func PrintLineByContainerName(containerName string) error {
	cont, exists := parseContainers(containerName, true)
	if exists {
		return PrintLineByContainerID(cont.ID)
	}
	return nil
}

func PrintLineByContainerID(containerID string) error {
	cont, _ := util.DockerClient.InspectContainer(containerID)
	return printLine(cont)
}

// this function populates the listing functions
func printLine(container *docker.Container) error {
	var line string
	// var n string
	tmp, _ := reflections.GetField(container, "Name")
	n := tmp.(string)
	Names := util.ContainerDisassemble(n)
	line = line + fmt.Sprintf("%-20s", Names.ShortName)
	line = line + fmt.Sprintf("%-10s", Names.Type)
	line = line + fmt.Sprintf("%-6d", Names.Number)

	if probablyHasDataContainer(container) {
		line = line + fmt.Sprintf("%-5s", "yes")
	} else {
		line = line + fmt.Sprintf("%-5s", "no")
	}

	line = line + formulatePortsOutput(container)
	line = line + fmt.Sprintf("%-20s", Names.FullName)
	line = line + "\n"
	return writeTemplate(container, line)
}

// this function is for parsing single variables
func printField(container interface{}, field string) error {
	var line string
	FieldCamel := camelize(field)
	f, _ := reflections.GetFieldKind(container, FieldCamel)
	switch f.String() {
	case "ptr":
		// we don't recurse into to gain a bit more control... this function will be rarely used and doesn't have to be perfectly parseable.
	case "map":
		line = fmt.Sprintf("{{ range $key, $val := .%v }}{{ $key }}->{{ $val }}\n{{ end }}", FieldCamel)
	case "slice":
		line = fmt.Sprintf("{{ range .%v }}{{ . }}\n{{ end }}", FieldCamel)
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

// ----------------------------------------------------------------------------
// Helpers

func probablyHasDataContainer(container *docker.Container) bool {
	eFolder := container.Volumes["/home/eris/.eris"]
	if eFolder != "" {
		if strings.Contains(eFolder, "_data") {
			return true
		}
	}
	return false
}

func formulatePortsOutput(container *docker.Container) string {
	var ports string
	for k, v := range container.NetworkSettings.Ports {
		if len(v) != 0 {
			ports = ports + fmt.Sprintf("%v:%v->%v ", v[0].HostIP, v[0].HostPort, k) // published ports
		} else {
			ports = ports + fmt.Sprintf("%v ", k) // exposed, but not published ports
		}
	}

	split := strings.Split(ports, ",")
	ports = ""
	sort.Sort(sort.StringSlice(split))
	for _, p := range split {
		ports = ports + p + " "
	}

	return ports
}

func camelize(field string) string {
	if !startsUp(field) {
		return snaker.SnakeToCamel(field)
	}
	return field
}

func writeTemplate(container interface{}, toParse string) error {
	tmpl, err := template.New("field").Parse(toParse)
	if err != nil {
		return err
	}

	if err = tmpl.Execute(util.GlobalConfig.Writer, container); err != nil {
		return err
	}

	return nil
}

func startsUp(field string) bool {
	return unicode.IsUpper([]rune(field)[0])
}
