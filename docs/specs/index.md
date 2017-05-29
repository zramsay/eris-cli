---

type:   docs
layout: single
title: "Tutorials | Specifications"
index_file: ""
path: "content/docs/specs"
menu:
  tutorials:
    weight: 10

---

## Specifications

<div class="note">
{{% data_sites rename_docs %}}
</div>


### Services Specification

Services are defined in **service definition files**. These reside on the host in `~/.monax/services`.

Service definition files are formatted using `toml`.

[Read the Services Specification &nbsp;<i class="fa fa-chevron-circle-right" aria-hidden="true"></i>](/docs/specs/services_specification)


### Jobs Specification

Jobs are a portion of the upcoming Monax package management tooling. They enable ease of automation of smart contract deployment and runtime configuration with special utilizations for Burrow, the Monax blockchain. All tasks can be run through a simple yaml file and cover a range of cases from interaction with smart contracts, assertion testing, name registry, querying the state of the chain, setting permissions on certain addresses, or sending transactions. 

Jobs are defined in **job definition files**.

Job definition files are formatted in `yaml` and default file is `epm.yaml`.

Examples of monax job definition files are available in the [jobs_fixtures](https://github.com/monax/monax/tree/master/tests/jobs_fixtures) directory.

[Read the Jobs Specification &nbsp;<i class="fa fa-chevron-circle-right" aria-hidden="true"></i>](/docs/specs/jobs_specification)



## [<i class="fa fa-chevron-circle-left" aria-hidden="true"></i> All Tutorials](/docs/)




