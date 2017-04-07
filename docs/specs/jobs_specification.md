---

type:   docs
layout: single
title: "Specifications | Jobs Specification"

---

## Jobs Specification

<div class="note">
	<em>Note: As of 2017, our product has been renamed from Eris to Monax. This documentation refers to an earlier version of the software prior to this name change (<= 0.16). Later versions of this documentation (=> 0.17) will change the <code>eris</code> command and <code>~/.eris</code> directory to <code>monax</code> and <code>~/.monax</code> respectively.</em>
</div>

Jobs are defined in **job definition files**.

Action definition files are formatted in `yaml` and default file is `epm.yaml`.

Examples of job definition files are available in the [jobs_fixtures directory](https://github.com/monax/cli/tree/master/tests/jobs_fixtures).

Each job will perform its required action and then it will save the result of its job in a variable which can be utilized by jobs later in the sequence using jobs' [variable specification](/docs/specs/variable_specification).

## Job Types

Jobs are performed as sequentially based on the order they are given in the jobs definition file. By default, `monax pkgs do` will perform the entire sequence of jobs which has been outlined in a given jobs file.

Job categories are categorize into:

* [transaction jobs](#transaction-jobs)
* [contracts jobs](#contracts-jobs)
* [test jobs](#test-jobs)
* [other jobs](#other-jobs)

// TODO clean this up (RJ)

For each job which is specified, EPM will parse the following information:

{{ insert_definition "package.go" "Jobs" }}

Each job must then specify one **and only one** key which will determine the `type` of job which `monax pkgs do` should run. It is not invalid to add additional jobs, but only one of the jobs will be ran by `monax pkgs do`. The jobs and their purposes are outlined in the Job struct.

{{ insert_definition "package.go" "Job" }}

### Transaction Jobs

Transaction jobs exposed through the package manager are available in the following job types:

* [send](#send-jobs): a transaction which sends tokens from one account to another
* [register](#register-jobs): register a name in the native name registry
* [permission](#permission-jobs): update an account's permissions or roles (must be sent from an account which has root permissions on the chain)
* [bond](#bond-jobs): a bonding transaction ("make me a part of the validator pool")
* [unbond](#unbond-jobs): an unbonding transaction ("I want to leave the validator pool")
* [rebond](#rebond-jobs): a rebonding transaction ("oops. I actually don't want to leave the validator pool. Please add me back.")

#### Send Jobs

The send job will parse the following information:

{{ insert_definition "jobs.go" "Send" }}

#### Register Jobs

The register job will parse the following information:

{{ insert_definition "jobs.go" "RegisterName" }}

#### Permission Jobs

The permission job will parse the following information:

{{ insert_definition "jobs.go" "Permission" }}

#### Bond Jobs

The bond job will parse the following information:

{{ insert_definition "jobs.go" "Bond" }}

#### Unbond Jobs

The unbond job will parse the following information:

{{ insert_definition "jobs.go" "Unbond" }}

#### Rebond Jobs

The rebond job will parse the following information:

{{ insert_definition "jobs.go" "Rebond" }}

### Contracts Jobs

Contracts jobs exposed through the package manager are available in the following job types:

* [deploy](#deploy-jobs): deploy a single contract
* [call](#call-jobs): send a transaction to a contract (can only be sent to existing contracts)

#### Deploy Jobs

The deploy job will parse the following information:

{{ insert_definition "jobs.go" "Deploy" }}

#### Call Jobs

The call job will parse the following information:

{{ insert_definition "jobs.go" "Call" }}

### Test Jobs

Test jobs exposed through `monax pkgs do` are available in the following job types:

* [query-account](#queryaccount-jobs): get information about an account on the blockchain
* [query-contract](#querycontract-jobs): perform a simulated call against a specific contract (usually used to trigger accessor functions which retrieve information from a return variable of a contract's function)
* [query-name](#queryname-jobs): get information about a registered name using burrow's name registry functionality
* [query-vals](#queryvals-jobs): get information about the validator set
* [assert](#assert-jobs): assert a relationship between a key and a value (useful for testing purposes to make sure everything deployed properly and/or is working as it should)

#### QueryAccount Jobs

The query-account job will parse the following information:

{{ insert_definition "jobs.go" "QueryAccount" }}

#### QueryContract Jobs

The query-contract job will parse the following information:

{{ insert_definition "jobs.go" "QueryContract" }}

#### QueryName Jobs

The query-name job will parse the following information:

{{ insert_definition "jobs.go" "QueryName" }}

#### QueryVals Jobs

The query-vals job will parse the following information:

{{ insert_definition "jobs.go" "QueryVals" }}

#### Assert Jobs

The assert job will parse the following information:

{{ insert_definition "jobs.go" "Assert" }}

### Other Jobs

Other jobs exposed through the package manager are available in the following job types:

* [account](#account-jobs): set the account to use
* [set](#set-jobs): set the value of a variable

#### Account Jobs

The account job will parse the following information:

{{ insert_definition "jobs.go" "Account" }}

#### Set Jobs

The set job will parse the following information:

{{ insert_definition "jobs.go" "Set" }}


## [<i class="fa fa-chevron-circle-left" aria-hidden="true"></i> All Specifications](/docs/specs/)
