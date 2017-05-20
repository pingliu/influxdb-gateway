package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/pingliu/influxdb-gateway/gateway"

	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

var (
	configFilePath string
	logFilePath    string
)

func init() {
	flag.StringVar(&configFilePath, "config-file-path", "/etc/influxdb-gateway.toml", "config file path")
	flag.StringVar(&logFilePath, "log-file-path", "/var/log/influxdb-gateway.log", "log file path")
	flag.Parse()

	log.SetOutput(&lumberjack.Logger{
		Filename:   logFilePath,
		MaxSize:    100,
		MaxBackups: 5,
		MaxAge:     7,
	})
}

func main() {
	c, err := gateway.LoadConfig(configFilePath)
	if err != nil {
		log.Fatal(err)
	}
	gateway, err := gateway.New(c)
	if err != nil {
		log.Fatal(err)
	}
	err = gateway.Open()
	if err != nil {
		log.Fatal(err)
	}

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM, syscall.SIGKILL)
	log.Println("Listening for signals")
	go func() {
		<-signalCh
		gateway.Close()
		log.Println("Signal received, shutdown...")
	}()

	select {}
}
