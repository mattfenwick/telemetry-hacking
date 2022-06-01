package cli

import (
	"github.com/mattfenwick/telemetry-hacking/pkg/queue"
	"github.com/mattfenwick/telemetry-hacking/pkg/server"
	"github.com/mattfenwick/telemetry-hacking/pkg/utils"
	"github.com/mattfenwick/telemetry-hacking/pkg/worker"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"log"
	"os"
)

func Run() {
	//version.RunVersionCommand()
	command := setupRootCommand()
	if err := errors.Wrapf(command.Execute(), "run root command"); err != nil {
		log.Fatalf("unable to run root command: %+v", err)
		os.Exit(1)
	}
}

type Flags struct {
	Verbosity string
}

func setupRootCommand() *cobra.Command {
	flags := &Flags{}

	command := &cobra.Command{
		Use:  "telemetry",
		Args: cobra.ExactArgs(0),
	}

	command.AddCommand(worker.Setup())
	command.AddCommand(queue.Setup())
	command.AddCommand(server.Setup())

	command.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		return utils.SetUpLogger(flags.Verbosity)
	}

	command.PersistentFlags().StringVarP(&flags.Verbosity, "verbosity", "v", "info", "log level; one of [info, debug, trace, warn, error, fatal, panic]")

	return command
}
