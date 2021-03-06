package middle

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
	Port           int
	JaegerURL      string
	BottomHost     string
	BottomPort     int
	BottomGRPCPort int
}

func Setup() *cobra.Command {
	var configPath string

	command := &cobra.Command{
		Use: "middle",
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
	tp, err := utils.SetUpTracerProvider(config.JaegerURL, "middle")
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
	queue := NewMiddle(stop, config.BottomHost, config.BottomPort, config.BottomGRPCPort)

	logrus.Infof("instantiated middle: %+v", queue)
	SetupHTTPServer(queue)

	addr := fmt.Sprintf(":%d", config.Port)
	logrus.Infof("starting HTTP server on port %d", config.Port)
	utils.DoOrDie(http.ListenAndServe(addr, nil))

	utils.DoOrDie(errors.Errorf("TODO"))
}
