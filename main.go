package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"github.com/pingliu/influxdb-gateway/gateway"

	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

var (
	configFilePath string
	logFilePath    string
	logger         zap.Logger
)

func init() {
	flag.StringVar(&configFilePath, "config-file-path", "/etc/influxdb-gateway.toml", "config file path")
	flag.StringVar(&logFilePath, "log-file-path", "/var/log/influxdb-gateway.log", "log file path")
	flag.Parse()

	logger = zap.New(
		zap.NewTextEncoder(),
		zap.Output(zap.AddSync(&lumberjack.Logger{
			Filename:   logFilePath,
			MaxSize:    100,
			MaxBackups: 5,
			MaxAge:     7,
		})),
	)
}

func main() {
	c, err := gateway.LoadConfig(configFilePath)
	if err != nil {
		logger.Fatal(err.Error())
	}
	gateway, err := gateway.New(c, logger)
	if err != nil {
		logger.Fatal(err.Error())
	}

	err = gateway.Open()
	if err != nil {
		logger.Fatal(err.Error())
	}

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM, syscall.SIGKILL)
	logger.Info("Listening for signals")
	select {
	case <-signalCh:
		gateway.Close()
		logger.Info("Signal received, shutdown...")
	}
}
