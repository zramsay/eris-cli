package commands

import (
	"github.com/eris-ltd/eris-cli/agent"
	"github.com/eris-ltd/eris-cli/util"

	"github.com/spf13/cobra"
)

var Agents = &cobra.Command{
	Use:   "agent",
	Short: "start an agent",
	Long: `start an agent
An agent is local server that, when started,  exposes three endpoints:

  /chains	=> list running chains on the host (GET)
  /download	=> download a tar'ed contract bundle (POST)
  /install	=> download and deploy and tar'ed bundle (POST)

The command is used to support the Eris Contracts Library Marketplace.

Please see the pull request for more information about using
the agent and its endpoints:

  https://github.com/eris-ltd/eris-cli/pull/632

The agent is stopped with ctrl+c.`,
	Run: func(cmd *cobra.Command, args []string) { cmd.Help() },
}

// Build the agent subcommand
func buildAgentsCommand() {
	Agents.AddCommand(agentStart)
}

var agentStart = &cobra.Command{
	Use:   "start",
	Short: "start the agent",
	Long:  `Start the agent. Stop the agent with Ctrl+C`,
	Run:   StartAgent,
}

func StartAgent(cmd *cobra.Command, args []string) {
	util.IfExit(ArgCheck(0, "eq", cmd, args))
	util.IfExit(agent.StartAgent(do))
}
