package util

import (
	"bytes"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"text/template"
	"unicode"

	"github.com/eris-ltd/eris-cli/config"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/oleiade/reflections"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/olekukonko/tablewriter"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/serenize/snaker"
)

// currently only used by `eris ls, eris services/chains ls`
// flags for listing functions do things their own way -> prevents testing clusterf*ck
// [zr] should struct be implemented throughout?
type Parts struct {
	ShortName   string //known & existing & running
	Type        string
	Running     bool
	FullName    string
	Number      int
	PortsOutput string
}

func PrintInspectionReport(cont *docker.Container, field string) error {
	switch field {
	case "line":
		parts, err := printLine(cont, false) //can only inspect a running container...?
		if err != nil {
			return err
		}
		logger.Printf("%s\n", strings.Join(parts, " "))
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

func PrintTableReport(typ string, existing, all bool) (string, error) {
	logger.Debugf("PrintTableReport Initialized =>\t%s", typ)

	var conts []*ContainerName
	if !all {
		conts = ErisContainersByType(typ, existing)
	}

	buf := new(bytes.Buffer)
	table := tablewriter.NewWriter(buf)
	table.SetHeader([]string{fmt.Sprintf("%s NAME", strings.ToUpper(typ)), "TYPE", "RUNNING", "CONTAINER NAME", "CONTAINER #", "PORTS"})

	if all { //get all the things
		parts, _ := AssembleTable(typ)
		for _, p := range parts {
			table.Append(formatLine(p))
		}
	} else {
		for _, c := range conts {
			n, _ := PrintLineByContainerName(c.FullName, existing)
			if typ == "chain" {
				head, _ := GetHead()
				if n[0] == head {
					n[0] = fmt.Sprintf("**  %s", n[0]) // TODO: colorize this when we settle on a lib
				}
			}
			table.Append(n)
		}
	}

	// Styling
	table.SetBorder(false)
	table.SetCenterSeparator(" ")
	table.SetColumnSeparator(" ")
	table.SetRowSeparator("-")
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.Render()

	return buf.String(), nil
}

func PrintLineByContainerName(containerName string, existing bool) ([]string, error) {
	cont, exists := ParseContainers(containerName, true)
	if exists {
		return PrintLineByContainerID(cont.ID, existing)
	}
	return nil, nil //fail silently
}

func PrintLineByContainerID(containerID string, existing bool) ([]string, error) {
	cont, err := DockerClient.InspectContainer(containerID)
	if err != nil {
		return nil, err
	}
	return printLine(cont, existing)
}

func PrintPortMappings(id string, ports []string) error {
	cont, err := DockerClient.InspectContainer(id)
	if err != nil {
		return err
	}

	exposedPorts := cont.NetworkSettings.Ports

	var minimalDisplay bool
	if len(ports) == 1 {
		minimalDisplay = true
	}

	// Display everything if no port's requested.
	if len(ports) == 0 {
		for exposed := range exposedPorts {
			ports = append(ports, string(exposed))
		}
	}

	// Replace plain port numbers without suffixes with both "/tcp" and "/udp" suffixes.
	// (For example, replace ["53"] in a slice with ["53/tcp", "53/udp"].)
	normalizedPorts := []string{}
	for _, port := range ports {
		if !strings.HasSuffix(port, "/tcp") && !strings.HasSuffix(port, "/udp") {
			normalizedPorts = append(normalizedPorts, port+"/tcp", port+"/udp")
		} else {
			normalizedPorts = append(normalizedPorts, port)
		}
	}

	for _, port := range normalizedPorts {
		for _, binding := range exposedPorts[docker.Port(port)] {
			hostAndPortBinding := fmt.Sprintf("%s:%s", binding.HostIP, binding.HostPort)

			// If only one port request, display just the binding.
			if minimalDisplay {
				logger.Printf("%s\n", hostAndPortBinding)
			} else {
				logger.Printf("%s -> %s\n", port, hostAndPortBinding)
			}
		}
	}

	return nil
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

	Names := ContainerDisassemble(n)

	parts := []string{Names.ShortName, Names.Type, running, Names.FullName, fmt.Sprintf("%d", Names.Number), formulatePortsOutput(container)}
	return parts, nil
}

// this function is for parsing single variables
func printField(container interface{}, field string) error {
	logger.Debugf("Inspecting field =>\t\t%s\n", field)
	var line string

	// We allow fields to be passed using dot syntax, but
	// we have to make sure all fields are Camelized
	lineSplit := strings.Split(field, ".")
	for n, f := range lineSplit {
		lineSplit[n] = camelize(f)
	}
	FieldCamel := strings.Join(lineSplit, ".")

	f, _ := reflections.GetFieldKind(container, FieldCamel)
	logger.Debugln("field type", f.String())
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
	return snaker.SnakeToCamel(field)
	if !startsUp(field) {
		return snaker.SnakeToCamel(field)
	}
	return field
}

func writeTemplate(container interface{}, toParse string) error {
	logger.Debugf("Template parsing =>\t\t%s", toParse)
	tmpl, err := template.New("field").Parse(toParse)
	if err != nil {
		return err
	}

	if err = tmpl.Execute(config.GlobalConfig.Writer, container); err != nil {
		return err
	}

	return nil
}

func startsUp(field string) bool {
	return unicode.IsUpper([]rune(field)[0])
}

//XXX moved from /perform/docker_run.go
func ParseContainers(name string, all bool) (docker.APIContainers, bool) {
	logger.Debugf("Parsing Containers =>\t\t%s:%t\n", name, all)
	containers := listContainers(all)

	r := regexp.MustCompile(name)

	if len(containers) != 0 {
		for _, container := range containers {
			for _, n := range container.Names {
				if r.MatchString(n) {
					logger.Debugf("Container Found =>\t\t%s\n", name)
					return container, true
					// } else {
					// logger.Debugf("No match =>\t\t\t%s:%v\n", name, container.Names)
				}
			}
		}
	}
	logger.Debugf("Container Not Found =>\t\t%s\n", name)
	return docker.APIContainers{}, false
}

func listContainers(all bool) []docker.APIContainers {
	var container []docker.APIContainers
	r := regexp.MustCompile(`\/eris_(?:service|chain|data)_(.+)_\d`)

	contns, _ := DockerClient.ListContainers(docker.ListContainersOptions{All: all})
	for _, con := range contns {
		for _, c := range con.Names {
			match := r.FindAllStringSubmatch(c, 1)
			if len(match) != 0 {
				container = append(container, con)
			}
		}
	}

	return container
}

//----------------------------------------------------------
//---------------------helpers for ls w/o flags-------------

//XXX test with multiple containers of same definition!
func AssembleTable(typ string) ([]Parts, error) {
	// []string
	known := GetGlobalLevelConfigFilesByType(typ, false) // definition files

	typ = strings.TrimSuffix(typ, "s") // :(
	// []*ContainerName
	contsR := ErisContainersByType(typ, false) //running
	contsE := ErisContainersByType(typ, true)  //existing

	if len(contsE) == 0 && len(contsR) == 0 && len(known) == 0 {
		return []Parts{}, nil
	}

	var myTable []Parts
	addedAlready := make(map[string]bool)

	if typ != "data" {
		for _, name := range contsR {
			part, _ := makePartFromContainer(name.FullName, false)
			addedAlready[part.ShortName] = true //has to come after because short name needed
			part.Running = true
			myTable = append(myTable, part)
		}
	}

	for _, name := range contsE {
		part, _ := makePartFromContainer(name.FullName, false)
		if addedAlready[part.ShortName] == true {
			continue
		} else {
			part.Running = false
			myTable = append(myTable, part)
		}
	}

	if typ != "data" {
		for _, name := range known {
			if addedAlready[name] == true { //name from known == part.ShortName
				continue
			}
			part, _ := makePartFromContainer(name, true)
			part.Running = false
			myTable = append(myTable, part)
		}
	}

	return myTable, nil
}

func formatLine(p Parts) []string {
	var running string
	if p.Running {
		running = "Yes"
	} else {
		running = "No"
	}
	number := fmt.Sprintf("%d", p.Number)

	//must match header
	part := []string{p.ShortName, p.Type, running, p.FullName, number, p.PortsOutput}

	if p.Type == "chain" {
		head, _ := GetHead()
		if part[0] == head {
			part[0] = fmt.Sprintf("**  %s", part[0]) // TODO: colorize this when we settle on a lib
		}
	}
	return part
}

func makePartFromContainer(name string, defs bool) (v Parts, err error) {
	if defs {
		v.ShortName = name // the only value needed for known
	} else {
		// this block pulls out functionality from
		// PrintLineByContainerName{Id} & printLine
		var contID *docker.Container
		cont, exists := ParseContainers(name, true)
		if exists {
			contID, err = DockerClient.InspectContainer(cont.ID)
			if err != nil {
				return Parts{}, err
			}
		}
		if err != nil {
			return Parts{}, err
		}
		tmp, err := reflections.GetField(contID, "Name")
		if err != nil {
			return Parts{}, err
		}

		n := tmp.(string)

		Names := ContainerDisassemble(n)

		v = Parts{
			ShortName: Names.ShortName,
			Type:      Names.Type,
			//Running: set in previous function
			FullName:    Names.FullName,
			Number:      Names.Number,
			PortsOutput: formulatePortsOutput(contID),
		}
	}
	return v, nil
}
