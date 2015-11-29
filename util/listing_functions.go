package util

import (
	"fmt"
	"strings"

	"github.com/eris-ltd/eris-cli/definitions"
)

func ListAll(do *definitions.Do, typ string) (err error) {
	quiet := do.Quiet
	var result string
	if do.All == true { //overrides all the functionality used for flags/tests to stdout a nice table
		result, err = PrintTableReport(typ, true, true) //when latter bool is true, former one will be ignored...
		if err != nil {
			return err
		}

		fmt.Println(result)
	} else {

		testing := len(do.Operations.Args) != 0 && do.Operations.Args[0] == "testing"

		var resK, resR, resE string

		if do.Known {
			if typ != "data" { //no definition files for datas
				if resK, err = ListKnown(typ); err != nil {
					return err
				}
			}
			do.Result = resK
		}
		if do.Running {
			if resR, err = ListRunningOrExisting(quiet, testing, false, typ); err != nil {
				return err
			}
			do.Result = resR
		}
		if do.Existing {
			if resE, err = ListRunningOrExisting(quiet, testing, true, typ); err != nil {
				return err
			}
			do.Result = resE
		}
	}
	return nil
}

//looks for definition files in ~/.eris/typ
func ListKnown(typ string) (result string, err error) {

	result = strings.Join(GetGlobalLevelConfigFilesByType(typ, false), "\n")

	if typ == "chains" {
		var chainsNew []string
		head, _ := GetHead()
		chns := GetGlobalLevelConfigFilesByType(typ, false)

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

//lists the containers running for a chain/service
//[zr] eventually remotes/actions
func ListRunningOrExisting(quiet, testing, existing bool, typ string) (result string, err error) {
	re := "Running"
	if existing {
		re = "Existing"
	}
	logger.Debugf("Asking Docker Client for the %s Containers. Quiet? %v\n", re, quiet)

	if quiet || testing {
		if typ == "services" {
			result = strings.Join(ServiceContainerNames(existing), "\n")
		}
		if typ == "chains" {
			result = strings.Join(ChainContainerNames(existing), "\n")
		}
		if typ == "data" {
			result = strings.Join(DataContainerNames(), "\n")
		}

	} else {
		if typ == "services" {
			logger.Debugf("List%sRaw:PrintTable =>\t%s:%v\n", re, "service", existing)
			result, _ = PrintTableReport("service", existing, false) //false is for All, dealt with somewhere else
		}
		if typ == "chains" {
			logger.Debugf("List%sRaw:PrintTable =>\t%s:%v\n", re, "chain", existing)
			result, _ = PrintTableReport("chain", existing, false)
		}
		if typ == "data" {
			result = strings.Join(DataContainerNames(), "\n")
		}
	}
	return result, nil
}
