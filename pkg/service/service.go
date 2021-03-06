package service

import (
	"time"

	"github.com/containerum/kube-client/pkg/model"
	"github.com/containerum/nodeMetrics/pkg/metrics"
	"github.com/containerum/nodeMetrics/pkg/metrics/prometheus"
	"github.com/containerum/nodeMetrics/pkg/service/handlers/cpu"
	"github.com/containerum/nodeMetrics/pkg/service/handlers/memory"
	"github.com/containerum/nodeMetrics/pkg/service/handlers/storage"
	"github.com/containerum/utils/httputil"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Config struct {
	InfluxAddr     string
	CadvisorAddr   string
	PrometheusAddr string

	Username string
	Password string
	DB       string
}

type Service struct {
	*gin.Engine
	config Config
}

func NewService(config Config, status *model.ServiceStatus) (*Service, error) {
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:     true,
		FullTimestamp:   true,
		TimestampFormat: time.RubyDate,
	})
	logrus.SetLevel(logrus.DebugLevel)
	var metricsProvider metrics.Metrics = prometheus.New(prometheus.Config{
		Addr: config.PrometheusAddr,
	})

	var server = gin.New()
	server.Use(gin.ErrorLogger())
	server.Use(gin.Recovery())

	server.GET("/status", httputil.ServiceStatus(status))

	var CPUmetrics = server.Group("/cpu")
	{
		CPUmetrics.GET("/current", cpu.Current(metricsProvider))
		CPUmetrics.GET("/history", cpu.History(metricsProvider))
		CPUmetrics.GET("/history/ws", cpu.HistoryWS(metricsProvider))
		CPUmetrics.GET("/history/nodes", cpu.NodesHistory(metricsProvider))
		CPUmetrics.GET("/history/nodes/ws", cpu.NodesHistoryWS(metricsProvider))
	}
	var memoryMetrics = server.Group("/memory")
	{
		memoryMetrics.GET("/current", memory.Current(metricsProvider))
		memoryMetrics.GET("/history", memory.History(metricsProvider))
		memoryMetrics.GET("/history/ws", memory.HistoryWS(metricsProvider))
		memoryMetrics.GET("/history/nodes", memory.NodeHistory(metricsProvider))
		memoryMetrics.GET("/history/nodes/ws", memory.NodesHistoryWS(metricsProvider))
	}
	var storageMetrics = server.Group("/storage")
	{
		storageMetrics.GET("/current", storage.Current(metricsProvider))
	}
	return &Service{
		Engine: server,
		config: config,
	}, nil
}
