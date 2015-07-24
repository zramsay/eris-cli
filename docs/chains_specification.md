# Chains Specification

Chains are defined in **chain definition files**. These reside on the host in `~/.eris/chains`.

Chain definition files may be formatted in any of the following formats:

* `json`
* `toml` (default)
* `yaml`

eris will marshal the following fields from chain definition files:

```go
// name of the chain
Name string `json:"name" yaml:"name" toml:"name"`
// chain_id of the chain
ChainID string `json:"chain_id" yaml:"chain_id" toml:"chain_id"`

// same fields as in the Service Struct/Service Specification
Service    *Service    `json:"service" yaml:"service" toml:"service"`
```
