# CHains Definition Files Specification

```
type Chain struct {
  Name     string `json:"name" yaml:"name" toml:"name"`
  Type     string `json:"type" yaml:"type" toml:"type"`
  Location string `json:"directory" yaml:"directory" toml:"directory"`
  Service  *Service
}
```

chain definitions: override service