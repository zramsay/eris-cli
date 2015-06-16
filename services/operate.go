package services

import (
	"fmt"
	"os"
	"strings"
	"strconv"

	"github.com/eris-ltd/eris-cli/perform"
	"github.com/eris-ltd/eris-cli/util"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/code.google.com/p/go-uuid/uuid"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/cobra"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/viper"
)

func Start(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		fmt.Println("No Service Given. Please rerun command with a known service.")
		os.Exit(1)
	}
  verbose := cmd.Flags().Lookup("verbose").Changed

	var service util.Service
	var conf = viper.New()

	conf.AddConfigPath(util.ServicesPath)
	conf.SetConfigName(args[0])
	conf.ReadInConfig()

	err := conf.Marshal(&service)
	if err != nil {
		// TODO: error handling
		fmt.Println(err)
		os.Exit(1)
	}

	// Services must be given an image. Flame out if they do not.
	if service.Image == "" {
		fmt.Println("An \"image\" field is required in the service definition file.")
		os.Exit(1)
	}

	// If no name use image name
	if service.Name == "" {
		if service.Image != "" {
			service.Name = strings.Replace(service.Image, "/", "_", -1)
		}
	}

  running := ListRunningRaw()
  if len(running) != 0 {
	  for _, srv := range running {
	    if srv == service.Name {
	      if verbose {
	        fmt.Println("Service already started. Skipping.")
	      }
	      return
	    }
	  }
  }

	// Give name a unique identifier
  containerNumber := 1 // tmp
	service.Name = "eris_service_" + service.Name + "_" + strings.Split(uuid.New(), "-")[0] + "_" + strconv.Itoa(containerNumber)

	perform.DockerRun(&service, verbose)
}

func Configure(cmd *cobra.Command, args []string) {

}

func Inspect(cmd *cobra.Command, args []string) {
	endpoint := "unix:///var/run/docker.sock"
	client, _ := docker.NewClient(endpoint)
	imgs, _ := client.ListImages(docker.ListImagesOptions{All: false})
	for _, img := range imgs {
		fmt.Println("ID: ", img.ID)
		fmt.Println("RepoTags: ", img.RepoTags)
		fmt.Println("Created: ", img.Created)
		fmt.Println("Size: ", img.Size)
		fmt.Println("VirtualSize: ", img.VirtualSize)
		fmt.Println("ParentId: ", img.ParentID)
	}
	// switch args[0] {
	// case "container":

	// }
}

func Logs(cmd *cobra.Command, args []string) {

}

func Kill(cmd *cobra.Command, args []string) {

}
