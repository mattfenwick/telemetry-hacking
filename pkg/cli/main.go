package cli

import (
	"github.com/mattfenwick/telemetry-hacking/pkg/utils"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
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

	command.AddCommand(setupWorkerCommand())
	command.AddCommand(setupQueueCommand())
	command.AddCommand(setupServerCommand())

	command.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		return utils.SetUpLogger(flags.Verbosity)
	}

	command.PersistentFlags().StringVarP(&flags.Verbosity, "verbosity", "v", "info", "log level; one of [info, debug, trace, warn, error, fatal, panic]")

	return command
}

type WorkerArgs struct {
	Type string
}

func setupWorkerCommand() *cobra.Command {
	var configPath string

	command := &cobra.Command{
		Use: "worker",
		Run: func(cmd *cobra.Command, positionalArgs []string) {
			runWorkerCommand(configPath)
		},
	}

	command.Flags().StringVar(&configPath, "config-path", "", "path to json config file")

	return command
}

func runWorkerCommand(configPath string) {
	args := WorkerArgs{}
	utils.DoOrDie(utils.ReadJsonFromFile(&args, configPath))
	logrus.Infof("worker args: %+v", args)

	utils.DoOrDie(errors.Errorf("TODO"))
}

type QueueArgs struct {
}

func setupQueueCommand() *cobra.Command {
	var configPath string

	command := &cobra.Command{
		Use: "queue",
		Run: func(cmd *cobra.Command, positionalArgs []string) {
			runQueueCommand(configPath)
		},
	}

	command.Flags().StringVar(&configPath, "config-path", "", "path to json config file")

	return command
}

func runQueueCommand(configPath string) {
	args := QueueArgs{}
	utils.DoOrDie(utils.ReadJsonFromFile(&args, configPath))
	logrus.Infof("queue args: %+v", args)

	utils.DoOrDie(errors.Errorf("TODO"))
}

type ServerArgs struct {
}

func setupServerCommand() *cobra.Command {
	var configPath string

	command := &cobra.Command{
		Use: "server",
		Run: func(cmd *cobra.Command, positionalArgs []string) {
			runServerCommand(configPath)
		},
	}

	command.Flags().StringVar(&configPath, "config-path", "", "path to json config file")

	return command
}

func runServerCommand(configPath string) {
	args := ServerArgs{}
	utils.DoOrDie(utils.ReadJsonFromFile(&args, configPath))
	logrus.Infof("server args: %+v", args)

	utils.DoOrDie(errors.Errorf("TODO"))
}
