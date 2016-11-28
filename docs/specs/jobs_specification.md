# Jobs Specification

Jobs are defined in **epm definition files**.

Action definition files may be formatted in any of the following formats:

* `json`
* `toml`
* `yaml` (default)

Examples of epm definition files are available in the [`tests/fixtures` directory](https://github.com/eris-ltd/eris-pm/tree/master/tests/fixtures).

Each job will perform its required action and then it will save the result of its job in a variable which can be utilized by jobs later in the sequence using eris:pm's [variable specification](variable_specification).

# Jobs

Jobs are performed as sequentially based on the order they are given in the epm definition file. By default EPM will perform the entire sequence of jobs which has been outlined in a given jobs file.

Job categories are categorize into:

* [transaction jobs](#txJobs);
* [contracts jobs](#contractsJobs);
* [test jobs](#testJobs); and
* [other jobs](#otherJobs).

For each job which is specified, EPM will parse the following information:

{{ insert_definition "package.go" "Jobs" }}

Each job must then specify one **and only one** key which will determine the `type` of job which eris:pm should run. It is not invalid to add additional jobs, but only one of the jobs will be ran by eris:pm. The jobs and their purposes are outlined in the Job struct.

{{ insert_definition "package.go" "Job" }}

## <a name="txJobs"></a>Transaction Jobs

Transaction jobs exposed through eris:pm are available in the following job types:

* [send](#sendJob): a transaction which sends tokens from one account to another
* [register](#registerJob): register a name in the native name registry
* [permission](#permJob): update an account's permissions or roles (must be sent from an account which has root permissions on the chain)
* [bond](#bondJob): a bonding transaction ("make me a part of the validator pool")
* [unbond](#unbondJob): an unbonding transaction ("I want to leave the validator pool")
* [rebond](#rebondJob): a rebonding transaction ("oops. I actually don't want to leave the validator pool. Please add me back.")

### <a name="sendJob"></a>Send Jobs

The send job will parse the following information:

{{ insert_definition "jobs.go" "Send" }}

### <a name="registerJob"></a>Register Jobs

The register job will parse the following information:

{{ insert_definition "jobs.go" "RegisterName" }}

### <a name="permJob"></a>Permission Jobs

The permission job will parse the following information:

{{ insert_definition "jobs.go" "Permission" }}

### <a name="bondJob"></a>Bond Jobs

The bond job will parse the following information:

{{ insert_definition "jobs.go" "Bond" }}

### <a name="unbondJob"></a>Unbond Jobs

The unbond job will parse the following information:

{{ insert_definition "jobs.go" "Unbond" }}

### <a name="rebondJob"></a>Rebond Jobs

The rebond job will parse the following information:

{{ insert_definition "jobs.go" "Rebond" }}

## <a name="contractsJobs"></a>Contracts Jobs

Contracts jobs exposed through eris:pm are available in the following job types:

* [deploy](#deployJob): deploy a single contract
* [call](#callJob): send a transaction to a contract (can only be sent to existing contracts)

### <a name="deployJob"></a>Deploy Jobs

The deploy job will parse the following information:

{{ insert_definition "jobs.go" "Deploy" }}

### <a name="callJob"></a>Call Jobs

The call job will parse the following information:

{{ insert_definition "jobs.go" "Call" }}

## <a name="testJobs"></a>Test Jobs

Test jobs exposed through eris:pm are available in the following job types:

* [query-account](#queryAccountJob): get information about an account on the blockchain
* [query-contract](#queryContractJob): perform a simulated call against a specific contract (usually used to trigger accessor functions which retrieve information from a return variable of a contract's function)
* [query-name](#queryNameJob): get information about a registered name using eris:db's name registry functionaltiy
* [query-vals](#queryValsJob): get information about the validator set
* [assert](#assertJob): assert a relationship between a key and a value (useful for testing purposes to make sure everything deployed properly and/or is working as it should)

### <a name="queryAccountJob"></a>QueryAccount Jobs

The query-account job will parse the following information:

{{ insert_definition "jobs.go" "QueryAccount" }}

### <a name="queryContractJob"></a>QueryContract Jobs

The query-contract job will parse the following information:

{{ insert_definition "jobs.go" "QueryContract" }}

### <a name="queryNameJob"></a>QueryName Jobs

The query-name job will parse the following information:

{{ insert_definition "jobs.go" "QueryName" }}

### <a name="queryValsJob"></a>QueryVals Jobs

The query-vals job will parse the following information:

{{ insert_definition "jobs.go" "QueryVals" }}

### <a name="assertJob"></a>Assert Jobs

The assert job will parse the following information:

{{ insert_definition "jobs.go" "Assert" }}

## <a name="otherJobs"></a>Other Jobs

Other jobs exposed through eris:pm are available in the following job types:

* [account](#accountJob): set the account to use
* [set](#setJob): set the value of a variable

### <a name="accountJob"></a>Account Jobs

The account job will parse the following information:

{{ insert_definition "jobs.go" "Account" }}

### <a name="setJob"></a>Set Jobs

The set job will parse the following information:

{{ insert_definition "jobs.go" "Set" }}
