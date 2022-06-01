package server

import (
	"context"
	"fmt"
	"github.com/mattfenwick/telemetry-hacking/pkg/queue"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"time"

	"github.com/mattfenwick/telemetry-hacking/pkg/utils"
)

type Config struct {
	Port      int
	JaegerURL string
	QueueHost string
	QueuePort int
}

func Setup() *cobra.Command {
	var configPath string

	command := &cobra.Command{
		Use: "server",
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
	logrus.Infof("server config: %+v", config)

	// start telemetry setup
	tp, err := utils.SetUpTracerProvider(config.JaegerURL, "worker")
	utils.DoOrDie(err)

	outerContext, cancel := context.WithCancel(context.Background())
	defer cancel()

	defer func(closedContext context.Context) {
		timedContext, timedCancel := context.WithTimeout(closedContext, time.Second*5)
		defer timedCancel()
		utils.DoOrDie(tp.Shutdown(timedContext))
	}(outerContext)
	// end telemetry setup

	queueClient := queue.NewClient(config.QueueHost, config.QueuePort)

	for i := 0; i < 10; i++ {
		logrus.Infof("issuing request %d", i)
		status, jobErr := queueClient.SubmitJob(&queue.JobRequest{
			JobId:    fmt.Sprintf("%d", i),
			Function: "um",
			Args:     []string{"qrs"},
		})
		logrus.Infof("received status, err: %+v, %+v", status, jobErr)
		time.Sleep(1 * time.Second)
	}

	state, err := queueClient.GetState()
	utils.DoOrDie(err)
	logrus.Infof("state: \n%+v\n", state)

	// TODO
	//server := NewServer()
	//
	//logrus.Infof("instantiated server: %+v", server)
	//SetupHTTPServer(server)
	//
	//addr := fmt.Sprintf(":%d", config.Port)
	//logrus.Infof("starting HTTP server on port %d", config.Port)
	//utils.DoOrDie(http.ListenAndServe(addr, nil))
	//
	//utils.DoOrDie(errors.Errorf("TODO"))
}
