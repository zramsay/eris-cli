package services

import (
  "fmt"
  "os"
  "strings"
  "strconv"

  "github.com/eris-ltd/eris-cli/perform"

  def "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/definitions"
  dir "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common"
  "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/cobra"
  "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/viper"
)

func Start(cmd *cobra.Command, args []string) {
  checkServiceGiven(args)
  StartServiceRaw(args[0], cmd.Flags().Lookup("verbose").Changed)
}

func Logs(cmd *cobra.Command, args []string) {
  checkServiceGiven(args)
  LogsServiceRaw(args[0], cmd.Flags().Lookup("verbose").Changed)
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

func LogsServiceRaw(servName string, verbose bool) {
  service := LoadServiceDefinition(servName)
  perform.DockerLogs(service, verbose)
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

func StartServiceByService(service *def.Service, verbose bool) {
  for _, srv := range service.ServiceDeps {
    go StartServiceRaw(srv, verbose)
  }
  perform.DockerRun(service, verbose)
}

func LogsServiceByService(service *def.Service, verbose bool) {
  perform.DockerLogs(service, verbose)
}

func KillServiceByService(service *def.Service, verbose bool) {
  for _, srv := range service.ServiceDeps {
    go KillServiceRaw(srv, verbose)
  }
  perform.DockerStop(service, verbose)
}

func LoadServiceDefinition(servName string) (*def.Service) {
  var service def.Service
  serviceConf := loadServiceDefinition(servName)
  marshalService(serviceConf, &service)

  checkServiceHasImage(&service)
  checkServiceHasName(&service)
  checkServiceHasDataContainer(serviceConf, &service)
  checkDataContainerHasName(&service)

  return &service
}

func loadServiceDefinition(servName string) *viper.Viper {
  var serviceConf = viper.New()

  serviceConf.AddConfigPath(dir.ServicesPath)
  serviceConf.SetConfigName(servName)
  serviceConf.ReadInConfig()

  return serviceConf
}

func marshalService(serviceConf *viper.Viper, service *def.Service) {
  err := serviceConf.Marshal(service)
  if err != nil {
    // TODO: error handling
    fmt.Println(err)
    os.Exit(1)
  }
}

func checkServiceGiven(args []string) {
  if len(args) == 0 {
    // TODO: betterly error handling
    fmt.Println("No Service Given. Please rerun command with a known service.")
    os.Exit(1)
  }
}

func checkServiceHasImage(service *def.Service) {
  // Services must be given an image. Flame out if they do not.
  if service.Image == "" {
    fmt.Println("An \"image\" field is required in the service definition file.")
    os.Exit(1)
  }
}

func checkServiceHasName(service *def.Service) {
  // If no name use image name
  if service.Name == "" {
    if service.Image != "" {
      service.Name = strings.Replace(service.Image, "/", "_", -1)
    }
  }

  containerNumber := 1 // tmp
  service.Name = "eris_service_" + service.Name + "_" + strconv.Itoa(containerNumber)
}

func checkServiceHasDataContainer(serviceConf *viper.Viper, service *def.Service) {
  // toml bools don't really marshal well
  if serviceConf.GetBool("data_container") {
    service.DataContainer = true
  }
}

func checkDataContainerHasName(service *def.Service) {
  service.DataContainerName = ""
  if service.DataContainer {
    dataSplit := strings.Split(service.Name, "_")
    dataSplit[1] = "data"
    service.DataContainerName = strings.Join(dataSplit, "_")
  }
}