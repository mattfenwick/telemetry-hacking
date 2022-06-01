package server

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"net/http"

	"github.com/mattfenwick/telemetry-hacking/pkg/utils"
)

type Config struct {
	Port int
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

	// TODO set up telemetry

	server := NewServer()

	logrus.Infof("instantiated server: %+v", server)
	SetupHTTPServer(server)

	addr := fmt.Sprintf(":%d", config.Port)
	logrus.Infof("starting HTTP server on port %d", config.Port)
	utils.DoOrDie(http.ListenAndServe(addr, nil))

	utils.DoOrDie(errors.Errorf("TODO"))
}
