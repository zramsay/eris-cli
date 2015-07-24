# Services Specification

Services are defined in **service definition files**. These reside on the host in `~/.eris/services`.

Service definition files may be formatted in any of the following formats:

* `json`
* `toml` (default)
* `yaml`

eris will marshal the following fields from service definition files:

```go
Name        string   `json:"name" yaml:"name" toml:"name"`
// id of the service
ServiceID   string   `json:"service_id" yaml:"service_id" toml:"service_id"`
// array of strings of other services which should be started prior to this service starting
ServiceDeps []string `json:"services", yaml:"services" toml:"services"`
// a chain which must be started prior to this service starting. can take a `$chain` string
// which would then be passed in via a command line flag
Chain       string   `json:"chain" yaml:"chain" toml:"chain"`

Service    *Service  `json:"service" yaml:"service" toml:"service"`
```

```go
// name of the service
Name        string   `json:"name" yaml:"name" toml:"name"`
// docker image used by the service
Image       string   `json:"image" yaml:"image" toml:"image"`
// whether eris should automagically handle a data container for this service
AutoData    bool     `json:"data_container" yaml:"data_container" toml:"data_container"`
// maps directly to docker cmd
Command     string   `json:"command" yaml:"command" toml:"command"`
// maps directly to docker links
Links       []string `json:"links" yaml:"links" toml:"links"`
// maps directly to docker ports
Ports       []string `json:"ports" yaml:"ports" toml:"ports"`
// maps directly do docker expose
Expose      []string `json:"expose" yaml:"expose" toml:"expose"`
// maps directly to docker volumes
Volumes     []string `json:"volumes" yaml:"volumes" toml:"volumes"`
// maps directly to docker volumes-from
VolumesFrom []string `json:"volumes_from" yaml:"volumes_from" toml:"volumes_from"`
// maps directly to docker environment
Environment []string `json:"environment" yaml:"environment" toml:"environment"`
// maps directly to docker env-file
EnvFile     []string `json:"env_file" yaml:"env_file" toml:"env_file"`
// maps directly to docker net
Net         string   `json:"net" yaml:"net" toml:"net"`
// maps directly to docker PID
PID         string   `json:"pid" yaml:"pid" toml:"pid"`
// maps directly to docker DNS
DNS         []string `json:"dns" yaml:"dns" toml:"dns"`
// maps directly to docker DNS-search
DNSSearch   []string `json:"dns_search" yaml:"dns_search" toml:"dns_search"`
// maps directly to docker workdir
WorkDir     string   `json:"work_dir" yaml:"work_dir" toml:"work_dir"`
// maps directly to docker entrypoint
EntryPoint  string   `json:"entry_point" yaml:"entry_point" toml:"entry_point"`
// maps directly to docker hostname
HostName    string   `json:"host_name" yaml:"host_name" toml:"host_name"`
// maps directly to docker domainname
DomainName  string   `json:"domain_name" yaml:"domain_name" toml:"domain_name"`
// maps directly to docker username
User        string   `json:"user" yaml:"user" toml:"user"`
// maps directly to docker cpu_shares
CPUShares   int64    `json:"cpu_shares" yaml:"cpu_shares" toml:"cpu_shares"`
// maps directly to docker mem_limit
MemLimit    int64    `json:"memory" yaml:"memory" toml:"memory"`
```

## Service Dependencies

Service dependencies are started by eris prior to the service itself starting.
