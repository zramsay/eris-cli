# eris changelog
## v0.16.0
This is a consolidation release that merges in two repositories (eris-cm & eris-pm), thereby eliminating their respective docker images and the many edge-case bugs that came with the spin-up-a-temporary-docker-container-to-execute-a-command approach previously taken. For both of these consolidations, the many additional small PRs required to sort everything out are not document here. In addition, both commands that used those repositories now run significantly faster.

#### Consolidation of eris-cm
- the `eris chains make` command no longer requires the `quay.io/eris/cm` docker image and, consequently, does not spin up temporary docker containers. [#1072](https://github.com/monax/cli/pull/1072) 
- a `--seeds-ip` flag has been added to `eris chains make` for setting the `seeds = "IP:PORT"` in the `config.toml`; a requirement for multi-node chains. [#1098](https://github.com/monax/cli/pull/1098)
- additional sample chain-types have been added on `eris init` to simplify answering the question "what combination of account-types should I use for my application". [#1097](https://github.com/monax/cli/pull/1097)
- a simplechain (one account) now uses the same directory structure as multi-node chains. [#1096](https://github.com/monax/cli/pull/1096)
- the `--unsafe` flag is now required for the `privKey` field in the `accounts.json` to be included (required by the javascript libraries). [#1235](https://github.com/monax/cli/pull/1235)

#### Consolidation of eris-pm
- the `eris pkgs do` command no longer requires the `quay.io/eris/pm` docker image and, consequently, does not spin up temporary docker containers. [#1083](https://github.com/monax/cli/pull/1083)
- the `quay.io/eris/compilers` docker image was greatly reduced in size and, consequently, local compilers is used by default for `eris pkgs do` rather than the remote service hosted by Monax. [#1201](https://github.com/monax/cli/pull/1201)
- linking for .sol binaries added. [#1197](https://github.com/monax/cli/pull/1197)

#### Additional improvements
- the solidity tutorial series has been added to this repo. [#1165](https://github.com/monax/cli/pull/1165)
- unused code for the deprecated `eris agent` command was removed, reducing the cognitive overheard to learning the codebase. [#1175](https://github.com/monax/cli/pull/1175)
- the eris-abi repository was unforked in favour of the go-ethereum abi. [#1176](https://github.com/monax/cli/pull/1176)
- this repo was renamed from `eris-cli` to `eris`, which required re-writing the import paths. [#1183](https://github.com/monax/cli/pull/1183)
- the `glide` tool is now used for dependency management. [#1177](https://github.com/monax/cli/pull/1177)
- variable names for the account-type permissions have been harmonized. [#1234](https://github.com/monax/cli/pull/1234)

#### Other repositories
- see [here](https://github.com/monax/burrow/blob/v0.16.0/CHANGELOG.md) for the `eris-db` changelog.
- see [this PR](https://github.com/monax/compilers/pull/121) for the compilers changelog.
- no changes were made to `eris-keys`. 
