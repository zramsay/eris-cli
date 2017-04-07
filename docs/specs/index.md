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
	<em>Note: As of 2017, our product has been renamed from Eris to Monax. This documentation refers to an earlier version of the software prior to this name change (<= 0.16). Later versions of this documentation (=> 0.17) will change the <code>eris</code> command and <code>~/.monax</code> directory to <code>monax</code> and <code>~/.monax</code> respectively.</em>
</div>

### Variables Specification

Variables can be used for nearly every jobs field (largely with the exception of nonce and wait).

[Read the Variables Specification &nbsp;<i class="fa fa-chevron-circle-right" aria-hidden="true"></i>](/docs/specs/variable_specification)


### Services Specification

Services are defined in **service definition files**. These reside on the host in `~/.monax/services`.

Service definition files are formatted using `toml`.

[Read the Services Specification &nbsp;<i class="fa fa-chevron-circle-right" aria-hidden="true"></i>](/docs/specs/services_specification)


### Jobs Specification

Jobs are defined in **job definition files**.

Action definition files are formatted in `yaml` and default file is `epm.yaml`.

Examples of job definition files are available in the jobs_fixtures directory.

[Read the Jobs Specification &nbsp;<i class="fa fa-chevron-circle-right" aria-hidden="true"></i>](/docs/specs/jobs_specification)



### Assert Jobs Specification

Asserts can be used to compare two "things". These "things" may be the result of two jobs or the result against one job against a baseline. (Indeed, it could be the comparison of two baselines but that wouldn't really get folks anywhere).

[Read the Assert Jobs Specification &nbsp;<i class="fa fa-chevron-circle-right" aria-hidden="true"></i>](/docs/specs/asserts_specification)



## [<i class="fa fa-chevron-circle-left" aria-hidden="true"></i> All Tutorials](/docs/)




