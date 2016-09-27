# Chains Specification

Chains are defined in **chain definition files**. These reside on the host in `~/.eris/chains`.

Chain definition files may be formatted in any of the following formats:

* `json`
* `toml` (default)
* `yaml`

eris will marshal the following fields from chain definition files:

{{ insert_definition "chain_definition.go" "ChainDefinition" }}
