package bottom

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"net/http"
	"time"

	"github.com/mattfenwick/telemetry-hacking/pkg/utils"
)

type Config struct {
	Port        int
	JaegerURL   string
	ThreadCount int
	GRPCPort    int
}

func Setup() *cobra.Command {
	var configPath string

	command := &cobra.Command{
		Use: "bottom",
		Run: func(cmd *cobra.Command, positionalArgs []string) {
			Run(configPath)
		},
	}

	command.Flags().StringVar(&configPath, "config-path", "", "path to json config file")

	return command
}

func Run(configPath string) {
	config := Config{}
	utils.DoOrDie(utils.ReadJsonFromFile(&config, configPath))
	logrus.Infof("queue config: %+v", config)

	// start telemetry setup
	tp, err := utils.SetUpTracerProvider(config.JaegerURL, "bottom")
	utils.DoOrDie(err)

	outerContext, cancel := context.WithCancel(context.Background())
	defer cancel()

	defer func(closedContext context.Context) {
		timedContext, timedCancel := context.WithTimeout(closedContext, time.Second*5)
		defer timedCancel()
		utils.DoOrDie(tp.Shutdown(timedContext))
	}(outerContext)
	// end telemetry setup

	stop := make(chan struct{})
	bottom := NewBottom(config.ThreadCount, stop)

	grpc, err := NewGRPCServer(config.GRPCPort, bottom)
	utils.DoOrDie(err)
	logrus.Infof("started grpc server: %+v", grpc)

	logrus.Infof("instantiated bottom: %+v", bottom)
	SetupHTTPServer(bottom)

	addr := fmt.Sprintf(":%d", config.Port)
	logrus.Infof("starting HTTP server on port %d", config.Port)
	utils.DoOrDie(http.ListenAndServe(addr, nil))

	utils.DoOrDie(errors.Errorf("TODO"))
}
