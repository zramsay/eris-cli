package jobs

import (
	//"encoding/json"
	"strings"

	"github.com/monax/cli/log"
	"github.com/monax/cli/util"

	"github.com/hyperledger/burrow/client"
	"github.com/hyperledger/burrow/keys"
)

// This is the jobset, the manager of all the job runners. It holds onto essential information for interacting
// with the chain that is passed to it from the initial Do struct from the CLI during the loading stage in
// the loaders package (See LoadJobs method). The main purpose of it is to maintain the ordering of job execution,
// and maintain awareness of the jobs that have been run and the results of those job runs.
type Jobs struct {
	// Chain and key specific variables
	Account       string            `json:"-"`
	NodeClient    client.NodeClient `json:"-"`
	KeyClient     keys.KeyClient    `json:"-"`
	ChainID       string            `json:"chain_ID"`
	PublicKey     string            `json:"-"`
	DefaultAddr   string            `json:"-"`
	DefaultAmount string            `json:"-"`
	DefaultGas    string            `json:"-"`
	DefaultFee    string            `json:"-"`
	// UI specific variables
	OutputFormat string   `json:"-"`
	DefaultSets  []string `json:"-"`
	Overwrite    bool     `json:"-"`
	//Path variables
	BinPath      string `json:"-"`
	AbiPath      string `json:"-"`
	ContractPath string `json:"-"`
	//Job variables
	Jobs       []*Job                 `mapstructure:"jobs" yaml:"jobs" json:"-"`
	JobMap     map[string]*JobResults `json:"output"`
	jobCounter int
}

// Returns an initialized empty jobset
func EmptyJobs() *Jobs {
	return &Jobs{}
}

// The main function out of the jobset, runs the jobs from the jobs config file in a sequential order
// checks for overwriting of the results of an old jobset if there is a repeat
func (jobs *Jobs) RunJobs() error {
	var jobNames []string
	if len(jobs.DefaultSets) >= 1 {
		jobs.defaultSetJobs()
	}
	if jobs.DefaultAddr != "" {
		jobs.defaultAddrJob()
	}
	/*var fileOutput []byte
	if output == "csv" {

	} else {
		fileOutput, err := json.Marshall(jobs)
	}*/
	for i, job := range jobs.Jobs {
		jobs.jobCounter = i
		// handle duplicate job names. Request user input for permission to overwrite.
		found, overwrite, at := checkForDuplicateQueryOverwrite(job.Name, jobNames, jobs.Overwrite)
		if found && !overwrite {
			continue
		} else if found && overwrite {
			//overwrite the name
			jobs.JobMap[jobNames[at]] = &JobResults{}
			jobNames = append(jobNames[:at], jobNames[at+1:]...)
		}

		jobNames = append(jobNames, job.Name)
		job.swapLegacyJob()
		results, err := job.beginJob(jobs)
		if err != nil {
			return err
		}
		jobs.JobMap[job.Name] = results
	}

	return nil
}

// The default address to work from with future jobs. Placed at the beginning of the jobset.
func (jobs *Jobs) defaultAddrJob() {
	oldJobs := jobs.Jobs

	newJob := []*Job{
		{
			Name: "defaultAddr",
			Account: &Account{
				Address: jobs.DefaultAddr,
			},
		},
	}

	jobs.Jobs = append(newJob, oldJobs...)
}

func (jobs *Jobs) defaultSetJobs() {
	oldJobs := jobs.Jobs

	newJobs := []*Job{}

	for _, setr := range jobs.DefaultSets {
		blowdUp := strings.Split(setr, "=")
		if blowdUp[0] != "" {
			newJobs = append(newJobs, &Job{
				Name: blowdUp[0],
				Set: &Set{
					Value: blowdUp[1],
				},
			})
		}
	}

	jobs.Jobs = append(newJobs, oldJobs...)
}

/*func (jobs *Jobs) marshalJSON() ([]byte, error) {

}*/

// this needs to change so that it isn't within the loop of the job functions and is rather gathered on a first round loop
// whereby the duplicate names are picked up and asked about prior to execution of the loop. This might make for a weird UI
// but there will be definite performance increases.
func checkForDuplicateQueryOverwrite(name string, jobNames []string, defaultOverwrite bool) (bool, bool, int) {
	var dup bool = false
	var index int = -1
	for i, checkForDup := range jobNames {
		if checkForDup == name {
			dup = true
			index = i
			break
		}
	}
	if dup {
		if defaultOverwrite {
			log.WithField("Overwriting job name", name)
		} else {
			overwriteWarning := "You are about to overwrite a previous job name, continue?"
			if util.QueryYesOrNo(overwriteWarning, []int{}...) == util.No {
				return true, false, index
			}
			return true, true, index
		}
	}
	return dup, defaultOverwrite, index
}

// This function handles post processing whereby the results are recorded.
// Post processing should be handled by taking in an error, if the error is nil and the current
// job counter == length of the contracts, then everything went off smoothly, record the entire job
// results based on the format that was requested. Otherwise there will
// be an error returned via this method, formatted and detailed, that will be returned but not
// before recording all of the job outputs up to this point.
/*func (jobs *Jobs) postProcess(makeOrBreak error) error {
	switch jobs.OutputFormat {
	case "csv":
		log.Info("Writing [epm.csv] to current directory")
		for _, job := range jobs.Jobs {
			if err := WriteJobResultCSV(job.JobName, job.JobResult); err != nil {
				return err
			}
		}
	case "json":
		log.Info("Writing [jobs_output.json] to current directory")
		results := make(map[string]string)
		for _, job := range do.Package.Jobs {
			results[job.JobName] = job.JobResult
		}
		return WriteJobResultJSON(results)
	}

	return nil
}*/

/* [rj] commenting out for the sake of unit testing loading
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
		if do.Overwrite == true && dup == true {
			log.WithField("Overwriting job name", job.JobName)
		} else if do.Overwrite == false && dup == true {
			overwriteWarning := "You are about to overwrite a previous job name, continue?"

			if util.QueryYesOrNo(overwriteWarning, []int{}...) == util.No {
				continue
			}
		}

		switch {
		// Util jobs
		case job.Job.Account != nil:
			announce(job.JobName, "Account")
			job.JobResult, err = SetAccountJob(job.Job.Account, do)
		case job.Job.Set != nil:
			announce(job.JobName, "Set")
			job.JobResult, err = SetValJob(job.Job.Set, do)

		// Transaction jobs
		case job.Job.Send != nil:
			announce(job.JobName, "Sent")
			job.JobResult, err = SendJob(job.Job.Send, do)
		case job.Job.RegisterName != nil:
			announce(job.JobName, "RegisterName")
			job.JobResult, err = RegisterNameJob(job.Job.RegisterName, do)
		case job.Job.Permission != nil:
			announce(job.JobName, "Permission")
			job.JobResult, err = PermissionJob(job.Job.Permission, do)
		case job.Job.Bond != nil:
			announce(job.JobName, "Bond")
			job.JobResult, err = BondJob(job.Job.Bond, do)
		case job.Job.Unbond != nil:
			announce(job.JobName, "Unbond")
			job.JobResult, err = UnbondJob(job.Job.Unbond, do)
		case job.Job.Rebond != nil:
			announce(job.JobName, "Rebond")
			job.JobResult, err = RebondJob(job.Job.Rebond, do)

		// Contracts jobs
		case job.Job.Deploy != nil:
			announce(job.JobName, "Deploy")
			job.JobResult, err = DeployJob(job.Job.Deploy, do)
		case job.Job.Call != nil:
			announce(job.JobName, "Call")
			job.JobResult, job.JobVars, err = CallJob(job.Job.Call, do)
			if len(job.JobVars) != 0 {
				for _, theJob := range job.JobVars {
					log.WithField("=>", fmt.Sprintf("%s,%s", theJob.Name, theJob.Value)).Info("Job Vars")
				}
			}
		// State jobs
		case job.Job.RestoreState != nil:
			announce(job.JobName, "RestoreState")
			job.JobResult, err = RestoreStateJob(job.Job.RestoreState, do)
		case job.Job.DumpState != nil:
			announce(job.JobName, "DumpState")
			job.JobResult, err = DumpStateJob(job.Job.DumpState, do)

		// Test jobs
		case job.Job.QueryAccount != nil:
			announce(job.JobName, "QueryAccount")
			job.JobResult, err = QueryAccountJob(job.Job.QueryAccount, do)
		case job.Job.QueryContract != nil:
			announce(job.JobName, "QueryContract")
			job.JobResult, job.JobVars, err = QueryContractJob(job.Job.QueryContract, do)
			if len(job.JobVars) != 0 {
				for _, theJob := range job.JobVars {
					log.WithField("=>", fmt.Sprintf("%s,%s", theJob.Name, theJob.Value)).Info("Job Vars")
				}
			}
		case job.Job.QueryName != nil:
			announce(job.JobName, "QueryName")
			job.JobResult, err = QueryNameJob(job.Job.QueryName, do)
		case job.Job.QueryVals != nil:
			announce(job.JobName, "QueryVals")
			job.JobResult, err = QueryValsJob(job.Job.QueryVals, do)
		case job.Job.Assert != nil:
			announce(job.JobName, "Assert")
			job.JobResult, err = AssertJob(job.Job.Assert, do)
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
}*/
