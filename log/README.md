# Monax Logger

Monax logging package is a fork of the [logrus](https://github.com/Sirupsen/logrus) logger with the `MonaxFormatter` set by default.

It also avoids logrus' shortcoming of not able to ignore logging levels set by the package's `SetLevel()` function.

## Generic Initialization

```
import "github.com/monax/monax/log"

log.SetLevel(log.WarnLevel)
if do.Verbose {
  log.SetLevel(log.InfoLevel)
} else if do.Debug {
  log.SetLevel(log.DebugLevel)
}

```

## Initialization for Tests

```
import log "github.com/monax/monax/log"

log.SetLevel(log.ErrorLevel)
// log.SetLevel(log.InfoLevel)
// log.SetLevel(log.DebugLevel)
```

## Usage

Simple logging at various levels:

```
log.Fatalf("Error changing directory: %v")
log.Error(err)
log.Warn("Failed to get that file. Sorry")
log.Info("Chain container does not exist")
log.Debug("Not requiring input. Proceeding")
``` 

Logging with a stand out item (unnamed). That item is most often a container, a chain, or a service name.

```
log.WithField("=>", chains.Name).Debug("Starting chain")
``` 

Logging with a named stand out parameter. 

```
log.WithField("args", do.Operations.Args).Info("Performing action (from tests)")
log.WithField("image", srv.Image).Debug("Container does not exist. Creating")
log.WithField("drop", dropped).Debug()
```

Logging with multiple standout parameters.

```
log.WithFields(log.Fields{
	"from": do.Name,
	"to":   do.NewName,
}).Info("Renaming action")
  
log.WithFields(log.Fields{
  	"=>":        servName,
	"existing#": containerExist,
	"running#":  containerRun,
}).Info("Checking number of containers for")
```
  
Standout items are sorted alphabetically with the unnamed parameter `=>` always the first. The order can be changed by splitting one log message into multiple log messages:
  
```
log.WithField("running#", containerRun).Info()
log.WithField("existing#", containerExist).Info()
log.WithField("=>", servName).Info("Checking number of containers for")
```

## Recommended Style

See the [CONTRIBUTING](https://github.com/monax/monax/blob/master/.github/CONTRIBUTING.md#errors-and-log-messages-style) document.
