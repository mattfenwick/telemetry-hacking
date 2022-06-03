package cli

import (
	"github.com/mattfenwick/telemetry-hacking/pkg/bottom"
	"github.com/mattfenwick/telemetry-hacking/pkg/middle"
	"github.com/mattfenwick/telemetry-hacking/pkg/top"
	"github.com/mattfenwick/telemetry-hacking/pkg/utils"
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

	command.AddCommand(bottom.Setup())
	command.AddCommand(middle.Setup())
	command.AddCommand(top.Setup())

	command.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		return utils.SetUpLogger(flags.Verbosity)
	}

	command.PersistentFlags().StringVarP(&flags.Verbosity, "verbosity", "v", "info", "log level; one of [info, debug, trace, warn, error, fatal, panic]")

	return command
}
