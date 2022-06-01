package cli

import (
	"context"
	"github.com/mattfenwick/telemetry-hacking/pkg/queue"
	"github.com/mattfenwick/telemetry-hacking/pkg/server"
	"github.com/mattfenwick/telemetry-hacking/pkg/utils"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel/trace"
	"log"
	"os"
	"time"
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
	command.AddCommand(queue.Setup())
	command.AddCommand(server.Setup())

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

	tp, err := utils.SetUpJaegerTracerProvider("http://localhost:14268/api/traces", "worker")
	utils.DoOrDie(err)

	outerContext, cancel := context.WithCancel(context.Background())
	defer cancel()

	defer func(closedContext context.Context) {
		timedContext, timedCancel := context.WithTimeout(closedContext, time.Second*5)
		defer timedCancel()
		utils.DoOrDie(tp.Shutdown(timedContext))
	}(outerContext)

	utils.RunOperation(outerContext, "test-span", func(span trace.Span) error {
		return errors.Errorf("TODO")
	})

	utils.DoOrDie(errors.Errorf("TODO"))
}
