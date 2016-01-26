package list

import (
	"fmt"
	"strings"

	log "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/util"
)

//looks for definition files in ~/.eris/typ
func ListKnown(typ string) (result string, err error) {

	result = strings.Join(util.GetGlobalLevelConfigFilesByType(typ, false), "\n")

	if typ == "chains" {
		var chainsNew []string
		head, _ := util.GetHead()
		chns := util.GetGlobalLevelConfigFilesByType(typ, false)

		for _, c := range chns {
			switch c {
			case "default":
				continue
			case head:
				chainsNew = append(chainsNew, fmt.Sprintf("*%s", c))
			default:
				chainsNew = append(chainsNew, fmt.Sprintf("%s", c))
			}
		}
		result = strings.Join(chainsNew, "\n")
	}
	return result, nil
}

//XXX chains and services only
func ListAll(do *definitions.Do, typ string) (err error) {
	quiet := do.Quiet
	var result string
	if do.All == true { //overrides all the functionality used for flags/tests to stdout a nice table
		resK, err := ListKnown(typ)
		if err != nil {
			return err
		}
		do.Result = resK //for testing but not rly needed
		knowns := strings.Split(resK, "\n")
		typs := fmt.Sprintf("The known %s on your host kind marmot:", typ)
		log.WithField("=>", knowns[0]).Warn(typs)
		knowns = append(knowns[:0], knowns[1:]...)
		for _, known := range knowns {
			log.WithField("=>", known).Warn()
		}

		result, err = PrintTableReport(typ, true, true) //when latter bool is true, former one will be ignored...
		if err != nil {
			return err
		}
		contType := fmt.Sprintf("Active %s containers:", typ)
		log.Warn(contType)
		log.Warn(result)
	} else {

		var resK, resR, resE string

		if do.Known {
			if resK, err = ListKnown(typ); err != nil {
				return err
			}
			do.Result = resK
		}
		if do.Running {
			if resR, err = ListRunningOrExisting(quiet, false, typ); err != nil {
				return err
			}
			do.Result = resR
		}
		if do.Existing {
			if resE, err = ListRunningOrExisting(quiet, true, typ); err != nil {
				return err
			}
			do.Result = resE
		}
	}
	return nil
}

// pulled out for simplicity; neither known or running
func ListDatas(do *definitions.Do) error {
	var result string
	var err error
	if !do.Quiet {
		result, err = PrintTableReport("data", true, true)
		if err != nil {
			return err
		}
		log.Warn(result)
	} else {
		result = strings.Join(util.DataContainerNames(), "\n")
		do.Result = result
		log.Warn("Active data containers:")
		log.Warn(result)
	}

	return nil
}

func ListActions(do *definitions.Do) error {
	actions, err := ListKnown("actions")
	if err != nil {
		return err
	}
	do.Result = actions //for testing but not rly needed
	knowns := strings.Split(actions, "\n")
	log.WithField("=>", knowns[0]).Warn("The known actions on your host kind marmot:")
	knowns = append(knowns[:0], knowns[1:]...)
	for _, known := range knowns {
		log.WithField("=>", known).Warn()
	}

	return nil
}

//lists the containers running for a chain/service
//[zr] eventually remotes/actions
func ListRunningOrExisting(quiet, existing bool, typ string) (result string, err error) {
	re := "Running"
	if existing {
		re = "Existing"
	}
	log.WithField("status", strings.ToLower(re)).Debug("Asking Docker to list containers")
	//gotta go
	if quiet {
		if typ == "services" {
			result = strings.Join(util.ServiceContainerNames(existing), "\n")
		}
		if typ == "chains" {
			result = strings.Join(util.ChainContainerNames(existing), "\n")
		}
	} else {
		if typ == "services" {
			log.WithField("=>", fmt.Sprintf("service:%v", strings.ToLower(re))).Debug("Printing table")
			result, _ = PrintTableReport("service", existing, false) //false is for All, dealt with somewhere else
		}
		if typ == "chains" {
			log.WithField("=>", fmt.Sprintf("chain:%v", strings.ToLower(re))).Debugf("Printing table")
			result, _ = PrintTableReport("chain", existing, false)
		}
	}
	return result, nil
}
