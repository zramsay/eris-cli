package commands

import (
	"fmt"

	"github.com/monax/monax/definitions"

	"github.com/spf13/cobra"
)

func buildFlag(cmd *cobra.Command, do *definitions.Do, flag, typ string) { //doesn't return anything; just sets the command
	switch flag {
	case "known":
		cmd.Flags().BoolVarP(&do.Known, "known", "k", false, fmt.Sprintf("list all the %s definition files that exist", typ))
	case "existing":
		cmd.Flags().BoolVarP(&do.Existing, "existing", "e", false, fmt.Sprintf("list all the all current containers which exist for a %s", typ))
	case "running":
		cmd.Flags().BoolVarP(&do.Running, "running", "r", false, fmt.Sprintf("list all the current containers which are running for a %s", typ))
	case "quiet":
		cmd.Flags().BoolVarP(&do.Quiet, "quiet", "q", false, "machine parsable output")
	case "force":
		cmd.Flags().BoolVarP(&do.Force, "force", "f", false, "kill the container instantly without waiting to exit") //why do we even have a timeout??
	case "timeout":
		cmd.Flags().UintVarP(&do.Timeout, "timeout", "t", 10, "manually set the timeout; overridden by --force")
	// TODO deduplicate? (notice the false/true)
	case "volumes":
		cmd.Flags().BoolVarP(&do.Volumes, "vol", "o", false, "remove volumes")
	case "rm-volumes":
		cmd.Flags().BoolVarP(&do.Volumes, "vol", "o", true, "remove volumes")
	case "rm":
		cmd.Flags().BoolVarP(&do.Rm, "rm", "r", false, "remove containers after stopping")
	case "data":
		cmd.Flags().BoolVarP(&do.RmD, "data", "x", false, "remove data containers after stopping")
	case "publish":
		cmd.PersistentFlags().BoolVarP(&do.Operations.PublishAllPorts, "publish", "p", false, "publish random ports")
	case "ports":
		cmd.PersistentFlags().StringVarP(&do.Operations.Ports, "ports", "", "", "reassign ports")
	case "interactive":
		cmd.Flags().BoolVarP(&do.Operations.Interactive, "interactive", "i", false, "interactive shell")
	case "pull":
		cmd.Flags().BoolVarP(&do.Pull, "pull", "p", false, fmt.Sprintf("pull an updated version of the %s's base service image from docker hub", typ))
	case "env":
		cmd.PersistentFlags().StringSliceVarP(&do.Env, "env", "e", nil, "multiple env vars can be passed using the KEY1=val1,KEY2=val2 syntax")
	case "links":
		cmd.PersistentFlags().StringSliceVarP(&do.Links, "links", "l", nil, "multiple containers can be linked can be passed using the KEY1:val1,KEY2:val2 syntax")
	case "follow":
		cmd.Flags().BoolVarP(&do.Follow, "follow", "f", false, "follow logs, like [tail -f]")
	case "tail":
		cmd.Flags().StringVarP(&do.Tail, "tail", "t", "150", "number of lines to show from end of logs")
	case "file":
		cmd.Flags().BoolVarP(&do.File, "file", "", false, fmt.Sprintf("remove %s definition file as well as %s container", typ, typ))
	case "services":
		cmd.Flags().StringSliceVarP(&do.ServicesSlice, "services", "s", []string{}, "comma separated list of services to start")
	case "init-dir":
		cmd.PersistentFlags().StringVarP(&do.Path, "init-dir", "", "", "a directory whose contents should be copied into the chain's main dir")
	}
}
