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
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/olekukonko/tablewriter"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/serenize/snaker"
)

func PrintInspectionReport(cont *docker.Container, field string) error {
	switch field {
	case "line":
		parts, err := printLine(cont)
		if err != nil {
			return err
		}
		logger.Println(strings.Join(parts, " "))
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

func PrintTableReport(typ string, running bool) error {
	conts := util.ErisContainersByType(typ, running)
	if len(conts) == 0 {
		return nil
	}

	table := tablewriter.NewWriter(logger.Writer)
	table.SetHeader([]string{"SERVICE NAME", "TYPE", "CONTAINER #", "PORTS", "CONTAINER NAME"})
	for _, c := range conts {
		n, _ := PrintLineByContainerName(c.FullName)
		table.Append(n)
	}

	// Styling
	table.SetBorder(false)
	table.SetCenterSeparator(" ")
	table.SetColumnSeparator(" ")
	table.SetRowSeparator("-")
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.Render()
	return nil
}

func PrintLineByContainerName(containerName string) ([]string, error) {
	cont, exists := parseContainers(containerName, true)
	if exists {
		return PrintLineByContainerID(cont.ID)
	}
	return nil, nil //fail silently
}

func PrintLineByContainerID(containerID string) ([]string, error) {
	cont, err := util.DockerClient.InspectContainer(containerID)
	if err != nil {
		return nil, err
	}
	return printLine(cont)
}

// this function populates the listing functions
func printLine(container *docker.Container) ([]string, error) {
	tmp, err := reflections.GetField(container, "Name")
	if err != nil {
		return nil, err
	}
	n := tmp.(string)

	Names := util.ContainerDisassemble(n)
	parts := []string{Names.ShortName, Names.Type, fmt.Sprintf("%d", Names.Number), formulatePortsOutput(container), Names.FullName}
	return parts, nil
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
