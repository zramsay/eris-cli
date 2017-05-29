---

type:   docs
layout: single
title: "Specifications | Services Specification"

---

## Services Specification

<div class="note">
{{% data_sites rename_docs %}}
</div>

Services are defined in **service definition files**. These reside on the host in `~/.monax/services`.

Service definition files are formatted using `toml`.

See the [source code](https://github.com/monax/monax/blob/master/definitions/service_definition.go) for more details on fields that will be marshalled. These largely mirror the fields used in `docker run/start`.

### Service Dependencies

Service dependencies are started by `monax` prior to the service itself starting.

### Linking to Chains

Linking to chains is done in one of two ways. For the CLI, you will give `monax services start` a `--chain` flag with the name of the chain you are wanting to start along with the services. Chains will be started prior to any services booting to make sure they are available to the linked service.

Chains can also be linked via the `chain` setting in the service definition file. This setting can take **either** a named chain, **or** a `$chain` **variable**. If you use the `$chain` variable then the linked chain will be either the flag given (which will take precedence), or the currently checked out chain. If there is no chain checked out and there is no chain identified by a flag, the command will fail.

### Linking to Other Services

In the service dependency section you will give the string in the following format `SERVICENAME:DOCKERNAME:CONNECTIONTYPE` where the following applies:

* `SERVICENAME` would be the name of the monax service you want to create a link to.
* `DOCKERNAME` would be what we tell docker the name is (usually this will be blank).
* `CONNECTIONTYPE` is the type of connection you want to make to the dependency service:
  * `a` (default) will create a docker link to the container and mount the container's volumes
  * `v` will mount the container's volumes
  * `l` will link to the container
  * `n` will do neither of the above


## [<i class="fa fa-chevron-circle-left" aria-hidden="true"></i> All Specifications](/docs/specs/)
