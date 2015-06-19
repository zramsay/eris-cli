package services

import (
  "fmt"
  "os"
  "strings"
  "strconv"

  "github.com/eris-ltd/eris-cli/perform"
  "github.com/eris-ltd/eris-cli/util"

  "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/cobra"
  "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/viper"
)

func Start(cmd *cobra.Command, args []string) {
  checkServiceGiven(args)
  StartServiceRaw(args[0], cmd.Flags().Lookup("verbose").Changed)
}

func Logs(cmd *cobra.Command, args []string) {

}

func Kill(cmd *cobra.Command, args []string) {
  checkServiceGiven(args)
  KillServiceRaw(args[0], cmd.Flags().Lookup("verbose").Changed)
}

func StartServiceRaw(servName string, verbose bool) {
  service := LoadServiceDefinition(servName)

  if IsServiceRunning(service) {
    if verbose {
      fmt.Println("Service already started. Skipping.")
    }
  } else {
    StartServiceByService(service, verbose)
  }
}

func KillServiceRaw(servName string, verbose bool) {
  service := LoadServiceDefinition(servName)

  if IsServiceRunning(service) {
    KillServiceByService(service, verbose)
  } else {
    if verbose {
      fmt.Println("Service not currently running. Skipping.")
    }
  }
}

func StartServiceByService(service *util.Service, verbose bool) {
  for _, srv := range service.ServiceDeps {
    go StartServiceRaw(srv, verbose)
  }
  perform.DockerRun(service, verbose)
}

func KillServiceByService(service *util.Service, verbose bool) {
  for _, srv := range service.ServiceDeps {
    go KillServiceRaw(srv, verbose)
  }
  perform.DockerStop(service, verbose)
}

func LoadServiceDefinition(servName string) (*util.Service) {
  var service util.Service
  var serviceConf = viper.New()

  serviceConf.AddConfigPath(util.ServicesPath)
  serviceConf.SetConfigName(servName)
  serviceConf.ReadInConfig()

  err := serviceConf.Marshal(&service)
  if err != nil {
    // TODO: error handling
    fmt.Println(err)
    os.Exit(1)
  }

  // toml bools don't really marshal well
  if serviceConf.GetBool("data_container") {
    service.DataContainer = true
  }

  checkServiceHasImage(&service)
  checkServiceHasName(&service)
  checkDataContainerHasName(&service)

  return &service
}

func checkServiceGiven(args []string) {
  if len(args) == 0 {
    // TODO: betterly error handling
    fmt.Println("No Service Given. Please rerun command with a known service.")
    os.Exit(1)
  }
}

func checkServiceHasImage(service *util.Service) {
  // Services must be given an image. Flame out if they do not.
  if service.Image == "" {
    fmt.Println("An \"image\" field is required in the service definition file.")
    os.Exit(1)
  }
}

func checkServiceHasName(service *util.Service) {
  // If no name use image name
  if service.Name == "" {
    if service.Image != "" {
      service.Name = strings.Replace(service.Image, "/", "_", -1)
    }
  }

  containerNumber := 1 // tmp
  service.Name = "eris_service_" + service.Name + "_" + strconv.Itoa(containerNumber)
}

func checkDataContainerHasName(service *util.Service) {
  service.DataContainerName = ""
  if service.DataContainer {
    dataSplit := strings.Split(service.Name, "_")
    dataSplit[1] = "data"
    service.DataContainerName = strings.Join(dataSplit, "_")
  }
}