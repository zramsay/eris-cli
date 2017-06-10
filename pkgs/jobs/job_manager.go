package jobs

import (
	"fmt"
	"strings"

	"github.com/monax/monax/definitions"
	"github.com/monax/monax/log"
	"github.com/monax/monax/util"
)

func RunJobs(do *definitions.Do) error {
	var err error
	var dup bool = false
	// ADD DefaultAddr and DefaultSet to jobs array....
	// These work in reverse order and the addendums to the
	// the ordering from the loading process is lifo
	if len(do.DefaultSets) >= 1 {
		defaultSetJobs(do)
	}

	if do.DefaultAddr != "" {
		defaultAddrJob(do)
	}

	for index, job := range do.Package.Jobs {
		for _, checkForDup := range do.Package.Jobs[0:index] {
			if checkForDup.JobName == job.JobName {
				dup = true
				break
			}
		}
		if do.Overwrite && dup {
			log.WithField("Overwriting job name", job.JobName)
		} else if !do.Overwrite && dup {
			overwriteWarning := "You are about to overwrite a previous job name, continue?"

			if util.QueryYesOrNo(overwriteWarning, []int{}...) == util.No {
				continue
			}
		}

		switch {
		// Util jobs
		case job.Account != nil:
			announce(job.JobName, "Account")
			job.JobResult, err = SetAccountJob(job.Account, do)
		case job.Set != nil:
			announce(job.JobName, "Set")
			job.JobResult, err = SetValJob(job.Set, do)

		// Transaction jobs
		case job.Send != nil:
			announce(job.JobName, "Sent")
			job.JobResult, err = SendJob(job.Send, do)
		case job.RegisterName != nil:
			announce(job.JobName, "RegisterName")
			job.JobResult, err = RegisterNameJob(job.RegisterName, do)
		case job.Permission != nil:
			announce(job.JobName, "Permission")
			job.JobResult, err = PermissionJob(job.Permission, do)
		case job.Bond != nil:
			announce(job.JobName, "Bond")
			job.JobResult, err = BondJob(job.Bond, do)
		case job.Unbond != nil:
			announce(job.JobName, "Unbond")
			job.JobResult, err = UnbondJob(job.Unbond, do)
		case job.Rebond != nil:
			announce(job.JobName, "Rebond")
			job.JobResult, err = RebondJob(job.Rebond, do)

		// Contracts jobs
		case job.Deploy != nil:
			announce(job.JobName, "Deploy")
			job.JobResult, err = DeployJob(job.Deploy, do)
		case job.Call != nil:
			announce(job.JobName, "Call")
			job.JobResult, job.JobVars, err = CallJob(job.Call, do)
			if len(job.JobVars) != 0 {
				for _, theJob := range job.JobVars {
					log.WithField("=>", fmt.Sprintf("%s,%s", theJob.Name, theJob.Value)).Info("Job Vars")
				}
			}
		// State jobs
		case job.RestoreState != nil:
			announce(job.JobName, "RestoreState")
			job.JobResult, err = RestoreStateJob(job.RestoreState, do)
		case job.DumpState != nil:
			announce(job.JobName, "DumpState")
			job.JobResult, err = DumpStateJob(job.DumpState, do)

		// Test jobs
		case job.QueryAccount != nil:
			announce(job.JobName, "QueryAccount")
			job.JobResult, err = QueryAccountJob(job.QueryAccount, do)
		case job.QueryContract != nil:
			announce(job.JobName, "QueryContract")
			job.JobResult, job.JobVars, err = QueryContractJob(job.QueryContract, do)
			if len(job.JobVars) != 0 {
				for _, theJob := range job.JobVars {
					log.WithField("=>", fmt.Sprintf("%s,%s", theJob.Name, theJob.Value)).Info("Job Vars")
				}
			}
		case job.QueryName != nil:
			announce(job.JobName, "QueryName")
			job.JobResult, err = QueryNameJob(job.QueryName, do)
		case job.QueryVals != nil:
			announce(job.JobName, "QueryVals")
			job.JobResult, err = QueryValsJob(job.QueryVals, do)
		case job.Assert != nil:
			announce(job.JobName, "Assert")
			job.JobResult, err = AssertJob(job.Assert, do)
		}

		if err != nil {
			return err
		}

	}

	postProcess(do)
	return nil
}

func announce(job, typ string) {
	log.Warn("\n*****Executing Job*****\n")
	log.WithField("=>", job).Warn("Job Name")
	log.WithField("=>", typ).Info("Type")
}

func defaultAddrJob(do *definitions.Do) {
	oldJobs := do.Package.Jobs

	newJob := &definitions.Job{
		JobName: "defaultAddr",
		Account: &definitions.Account{
			Address: do.DefaultAddr,
		},
	}

	do.Package.Jobs = append([]*definitions.Job{newJob}, oldJobs...)
}

func defaultSetJobs(do *definitions.Do) {
	oldJobs := do.Package.Jobs

	newJobs := []*definitions.Job{}

	for _, setr := range do.DefaultSets {
		blowdUp := strings.Split(setr, "=")
		if blowdUp[0] != "" {
			newJobs = append(newJobs, &definitions.Job{
				JobName: blowdUp[0],
				Set: &definitions.SetJob{
					Value: blowdUp[1],
				},
			})
		}
	}

	do.Package.Jobs = append(newJobs, oldJobs...)
}

func postProcess(do *definitions.Do) error {
	// check do.YAMLPath and do.DefaultOutput
	// get the epm.yaml
	var yaml string
	yamlName := strings.LastIndexByte(do.YAMLPath, '.')
	if yamlName >= 0 {
		yaml = do.YAMLPath[:yamlName]
	} else {
		return fmt.Errorf("invalid jobs file path (%s)", do.YAMLPath)
	}

	// if do.YAMLPath is not default and do.DefaultOutput is default, over-ride do.DefaultOutput
	if yaml != "epm" && do.DefaultOutput == "epm.output.json" {
		do.DefaultOutput = fmt.Sprintf("%s.output.json", yaml)
	}

	log.Warn(fmt.Sprintf("Writing [%s] to current directory", do.DefaultOutput))
	results := make(map[string]string)
	for _, job := range do.Package.Jobs {
		results[job.JobName] = job.JobResult
	}
	return WriteJobResultJSON(results, do.DefaultOutput)

	return nil
}
