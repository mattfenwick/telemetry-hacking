package top

import (
	"context"
	"fmt"
	"github.com/mattfenwick/telemetry-hacking/pkg/bottom"
	"github.com/mattfenwick/telemetry-hacking/pkg/middle"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	"time"

	"github.com/mattfenwick/telemetry-hacking/pkg/utils"
)

type Config struct {
	Port       int
	JaegerURL  string
	MiddleHost string
	MiddlePort int
}

func Setup() *cobra.Command {
	var configPath string

	command := &cobra.Command{
		Use: "top",
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
	tp, err := utils.SetUpTracerProvider(config.JaegerURL, "top")
	utils.DoOrDie(err)

	outerContext, cancel := context.WithCancel(context.Background())
	defer cancel()

	defer func(closedContext context.Context) {
		timedContext, timedCancel := context.WithTimeout(closedContext, time.Second*5)
		defer timedCancel()
		utils.DoOrDie(tp.Shutdown(timedContext))
	}(outerContext)
	// end telemetry setup

	middleClient := middle.NewClient(config.MiddleHost, config.MiddlePort)

	requests := []*bottom.Function{
		{Name: "+"},
		{Name: "*"},
		{Name: "sleep", Args: []int{3111, 2111, 1111}},
		{Name: "+", Args: []int{32}},
		{Name: "*", Args: []int{32}},
		{Name: "sleep", Args: []int{2468}},
		{Name: "+", Args: []int{32, 45, 121, 18}},
		{Name: "*", Args: []int{32, 45, 121, 18}},
		{Name: "sleep", Args: []int{32, 2113}},
		{Name: "+", Args: []int{333, 444, 555}},
		{Name: "*", Args: []int{333, 444, 555}},
		{Name: "sleep"},
	}

	group, errorGroupContext := errgroup.WithContext(outerContext)
	for i := 0; i < len(requests); i++ {
		j := i
		group.Go(func() error {
			logrus.Infof("issuing request %d", j)
			result, jobErr := middleClient.SubmitJob(errorGroupContext, &middle.JobRequest{
				JobId:    fmt.Sprintf("%d", j),
				Function: requests[j].Name,
				Args:     requests[j].Args,
			})
			logrus.Infof("received status, err: %+v, %+v", result, jobErr)
			return nil
		})
	}
	_ = group.Wait()

	//state, err := middleClient.GetState()
	//utils.DoOrDie(err)
	//logrus.Infof("state: \n%+v\n", state)

	// TODO
	//server := NewTop()
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
