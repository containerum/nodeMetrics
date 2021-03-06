package main

import (
	"flag"
	"os"

	"github.com/containerum/kube-client/pkg/model"
	"github.com/containerum/nodeMetrics/pkg/service"
	"github.com/octago/sflags/gen/gflag"
	"github.com/sirupsen/logrus"
)

type Config struct {
	Help        bool `flag:"help h"`
	ServingAddr string
	service.Config
}

var version string

func main() {
	var config = Config{
		ServingAddr: "localhost:8090",
		Config: service.Config{
			DB:             "kubernetes",
			InfluxAddr:     "http://192.168.88.210:8086",
			CadvisorAddr:   "http://192.168.88.210:31314",
			PrometheusAddr: "http://192.168.88.210:9090",
		},
	}
	if err := gflag.ParseToDef(&config); err != nil {
		panic(err)
	}
	flag.Parse()
	if config.Help {
		flag.Usage()
		return
	}

	status := model.ServiceStatus{
		Name:     "node-metrics",
		Version:  version,
		StatusOK: true,
	}

	service, err := service.NewService(config.Config, &status)
	if err != nil {
		logrus.WithError(err).Errorf("unable to start service")
		os.Exit(1)
	}
	if err := service.Run(config.ServingAddr); err != nil {
		logrus.WithError(err).Errorf("error while service execution")
		os.Exit(1)
	}
}
