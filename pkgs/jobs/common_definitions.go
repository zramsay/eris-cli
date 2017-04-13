package jobs

import (
	"fmt"
	"reflect"

	"github.com/monax/cli/log"
)

type JobsRunner interface {
	PreProcess(*Jobs) error
	Execute(*Jobs) (*JobResults, error)
}

type Job struct {
	// Name of the job
	Name string `mapstructure:"name" yaml:"name" json:"name"`
	///
	/// UTILITY JOBS
	///
	// Sets/Resets the primary account to use
	Account *Account `mapstructure:"account" yaml:"account"`
	// Set an arbitrary value
	Set *Set `mapstructure:"set" yaml:"set"`
	///
	/// CONTRACT JOBS
	///
	// Compile files
	Compile *Compile `mapstructure:"compile" yaml:"compile"`
	// Contract compile and send to the chain functions
	Deploy *Deploy `mapstructure:"deploy" yaml:"deploy"`
	// Sends a transaction to a contract. Will utilize abi under the hood to perform all of the heavy lifting
	Call *Call `mapstructure:"call" yaml:"call"`
	///
	/// TRANSACTION JOBS
	///
	// Send tokens from one account to another
	Send *Send `mapstructure:"send" yaml:"send"`
	// Utilize burrow's native name registry to register a name
	RegisterName *RegisterName `mapstructure:"register" yaml:"register"`
	// Sends a transaction which will update the permissions of an account. Must be sent from an account which
	// has root permissions on the blockchain (as set by either the genesis.json or in a subsequence transaction)
	Permission *Permission `mapstructure:"permission" yaml:"permission"`
	///
	/// TESTING JOBS
	///
	// Sends a "simulated call" to a contract. Predominantly used for accessor functions ("Getters" within contracts)
	QueryContract *QueryContract `mapstructure:"query-contract" yaml:"query-contract"`
	// Queries information from an account.
	QueryAccount *QueryAccount `mapstructure:"query-account" yaml:"query-account"`
	// Queries information about a name registered with burrow's native name registry
	QueryName *QueryName `mapstructure:"query-name" yaml:"query-name"`
	// Queries information about the validator set
	QueryVals *QueryVals `mapstructure:"query-vals" yaml:"query-vals"`
	// Makes and assertion (useful for testing purposes)
	Assert *Assert `mapstructure:"assert" yaml:"assert"`
	/// TO IMPLEMENT - WILL ONLY SLOW THIS DOWN IF LEFT UNCOMMENTED
	// Sends a bond transaction
	//Bond *Bond `mapstructure:"bond" yaml:"bond"`
	// Sends an unbond transaction
	//Unbond *Unbond `mapstructure:"unbond" yaml:"unbond"`
	// Sends a rebond transaction
	//Rebond *Rebond `mapstructure:"rebond" yaml:"rebond"`
	// Wrapper for mintdump dump. WIP
	//DumpState *DumpState `mapstructure:"dump-state" yaml:"dump-state"`
	// Wrapper for mintdum restore. WIP
	//RestoreState *RestoreState `mapstructure:"restore-state" yaml:"restore-state"`
	// Legacy job field soon to be deprecated
	Legacy *LegacyJob `mapstructure:"job" yaml:"job"`
	// Results of the job
	Results JobResults
}

func (job *Job) getType() (JobsRunner, error) {
	jobType := reflect.ValueOf(job)
	jobValue := jobType.Elem()
	//iterate through struct field and execute non nil and non name fields
	//break after executing so that we don't run into other fields
	for i := 1; i < jobValue.NumField(); i++ {
		field := jobValue.Field(i)
		if ptr := field.Pointer(); ptr != 0 {
			return field.Interface().(JobsRunner), nil
		}
	}
	return nil, fmt.Errorf("Could not find a job to execute.")
}

func (job *Job) announce(jobType JobsRunner) {
	log.Warn("\n*****Executing Job*****\n")
	log.WithField("=>", job.Name).Warn("Job Name")
	typ := fmt.Sprintf("%T", jobType)
	log.WithField("=>", typ).Info("Type")
}

func (job *Job) beginJob(jobs *Jobs) (*JobResults, error) {
	var jobType JobsRunner
	jobType, err := job.getType()
	if err != nil {
		return &JobResults{}, err
	}
	job.announce(jobType)
	err = jobType.PreProcess(jobs)
	if err != nil {
		return &JobResults{}, err
	}
	results, err := jobType.Execute(jobs)
	if err != nil {
		return &JobResults{}, err
	}
	return results, nil
}

func (job *Job) MarshalJSON() ([]byte, error) {
	type Alias Job
	return json.Marshal(&struct {
		LastSeen int64 `json:"lastSeen"`
		*Alias
	}{
		LastSeen: u.LastSeen.Unix(),
		Alias:    (*Alias)(job),
	})
}

type Type struct {
	StringResult string `mapstructure:"result" json:"result"`
	ActualResult interface{}
}

type JobResults struct {
	FullResult   Type            `mapstructure:"full_result" json:"full_result"`
	NamedResults map[string]Type `mapstructure:"named_results" json:"named_results"`
}

type LegacyJob struct {
	///
	/// UTILITY JOBS
	///
	// Sets/Resets the primary account to use
	Account *Account `mapstructure:"account" yaml:"account"`
	// Set an arbitrary value
	Set *Set `mapstructure:"set" yaml:"set"`
	///
	/// CONTRACT JOBS
	///
	// Compile files
	Compile *Compile `mapstructure:"compile" yaml:"compile"`
	// Contract compile and send to the chain functions
	Deploy *Deploy `mapstructure:"deploy" yaml:"deploy"`
	// Sends a transaction to a contract. Will utilize abi under the hood to perform all of the heavy lifting
	Call *Call `mapstructure:"call" yaml:"call"`
	///
	/// TRANSACTION JOBS
	///
	// Send tokens from one account to another
	Send *Send `mapstructure:"send" yaml:"send"`
	// Utilize burrow's native name registry to register a name
	RegisterName *RegisterName `mapstructure:"register" yaml:"register"`
	// Sends a transaction which will update the permissions of an account. Must be sent from an account which
	// has root permissions on the blockchain (as set by either the genesis.json or in a subsequence transaction)
	Permission *Permission `mapstructure:"permission" yaml:"permission"`
	///
	/// TESTING JOBS
	///
	// Sends a "simulated call" to a contract. Predominantly used for accessor functions ("Getters" within contracts)
	QueryContract *QueryContract `mapstructure:"query-contract" yaml:"query-contract"`
	// Queries information from an account.
	QueryAccount *QueryAccount `mapstructure:"query-account" yaml:"query-account"`
	// Queries information about a name registered with burrow's native name registry
	QueryName *QueryName `mapstructure:"query-name" yaml:"query-name"`
	// Queries information about the validator set
	QueryVals *QueryVals `mapstructure:"query-vals" yaml:"query-vals"`
	// Makes and assertion (useful for testing purposes)
	Assert *Assert `mapstructure:"assert" yaml:"assert"`
	/// TO IMPLEMENT - WILL ONLY SLOW THIS DOWN IF LEFT UNCOMMENTED
	// Sends a bond transaction
	//Bond *Bond `mapstructure:"bond" yaml:"bond"`
	// Sends an unbond transaction
	//Unbond *Unbond `mapstructure:"unbond" yaml:"unbond"`
	// Sends a rebond transaction
	//Rebond *Rebond `mapstructure:"rebond" yaml:"rebond"`
	// Wrapper for mintdump dump. WIP
	//DumpState *DumpState `mapstructure:"dump-state" yaml:"dump-state"`
	// Wrapper for mintdum restore. WIP
	//RestoreState *RestoreState `mapstructure:"restore-state" yaml:"restore-state"`
}

func (job *LegacyJob) deprecationNotice() {
	log.Warn("Deprecation Warning: the `job` field is no longer required and is soon to be deprecated. Please rid it from your file and use the new format.")
}

func (job *Job) swapLegacyJob() {
	legacy := job.Legacy
	if legacy != nil {
		job.Account = legacy.Account
		job.Set = legacy.Set
		job.Deploy = legacy.Deploy
		job.Send = legacy.Send
		job.RegisterName = legacy.RegisterName
		job.Permission = legacy.Permission
		//job.Bond = legacy.Bond
		//job.Unbond = legacy.Unbond
		//job.Rebond = legacy.Rebond
		//job.DumpState = legacy.DumpState
		//job.RestoreState = legacy.RestoreState
		job.QueryAccount = legacy.QueryAccount
		job.QueryContract = legacy.QueryContract
		job.Assert = legacy.Assert
		job.Call = legacy.Call
		legacy.deprecationNotice()
		job.Legacy = nil
	}
	return
}
