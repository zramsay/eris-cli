package list

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/eris-ltd/eris-cli/util"

	log "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/oleiade/reflections"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/olekukonko/tablewriter"
)

func PrintTableReport(typ string, existing, all bool) (string, error) {
	log.WithField("type", typ).Debug("Table report initialized")

	var conts []*util.ContainerName
	if !all {
		conts = util.ErisContainersByType(typ, existing)
	}
	// "MACHINE" is placeholder
	header := []string{"NAME", "MACHINE", "RUNNING", "CONTAINER NAME", "PORTS"}
	if err := util.CheckParts(header); err != nil {
		log.Error(err) // err is silenced by some funcs
		return "", err
	}

	buf := new(bytes.Buffer)
	table := tablewriter.NewWriter(buf)
	//name set by logger instead
	table.SetHeader(header)

	if all { //get all the things
		parts, _ := AssembleTable(typ)
		for _, p := range parts {
			table.Append(formatLine(p))
		}
	} else {
		for _, c := range conts {
			n, _ := PrintLineByContainerName(c.FullName, existing)
			if typ == "chain" {
				head, _ := util.GetHead()
				if n[0] == head {
					n[0] = fmt.Sprintf("**  %s", n[0])
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

type Parts struct {
	ShortName   string //known & existing & running
	Machine     string //TODO
	Running     bool
	FullName    string
	PortsOutput string
}

func PrintLineByContainerName(containerName string, existing bool) ([]string, error) {
	cont0, exists := util.ParseContainers(containerName, true)
	if exists {
		cont1, running := util.ParseContainers(containerName, false)
		if running && (cont0.ID == cont1.ID) {
			return util.PrintLineByContainerID(cont0.ID, false) //will print "RUNNING" = Yes when asked for existing/running conts

		} else {
			return util.PrintLineByContainerID(cont0.ID, true) // will print "RUNNING" = No when asked for existing conts
		}
	}
	return nil, nil //fail silently
}

//----------------------------------------------------------
//---------------------helpers for ls w/o flags-------------

//XXX test with multiple containers of same definition!
func AssembleTable(typ string) ([]Parts, error) {

	typ = strings.TrimSuffix(typ, "s") // :(
	// []*ContainerName
	contsR := util.ErisContainersByType(typ, false) //running
	contsE := util.ErisContainersByType(typ, true)  //existing

	if len(contsE) == 0 && len(contsR) == 0 {
		return []Parts{}, nil
	}

	var myTable []Parts
	addedAlready := make(map[string]bool)

	for _, name := range contsR {
		part, _ := makePartFromContainer(name.FullName)
		addedAlready[part.ShortName] = true //has to come after because short name needed
		part.Running = true
		myTable = append(myTable, part)
	}

	for _, name := range contsE {
		part, _ := makePartFromContainer(name.FullName)
		if addedAlready[part.ShortName] == true {
			continue
		} else {
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

	//must match header
	part := []string{p.ShortName, "", running, p.FullName, p.PortsOutput}
	if err := util.CheckParts(part); err != nil {
		log.Error(err)
		return []string{}
	}

	return part
}

func makePartFromContainer(name string) (v Parts, err error) {
	// this block pulls out functionality from
	// PrintLineByContainerName{Id} & printLine
	var contID *docker.Container
	cont, exists := util.ParseContainers(name, true)
	if exists {
		contID, err = util.DockerClient.InspectContainer(cont.ID)
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

	Names := util.ContainerDisassemble(n)

	v = Parts{
		ShortName: Names.ShortName,
		//Running: set in previous function
		FullName:    Names.FullName,
		PortsOutput: util.FormulatePortsOutput(contID),
	}
	return v, nil
}
