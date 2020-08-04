package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/rs/zerolog"
)

var logger zerolog.Logger

func main() {
	prog := os.Args[0]
	logFileName := prog + ".log"
	logFile, err := os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal("could not open log file: " + err.Error())
		os.Exit(1)
	}

	logger = zerolog.New(logFile).With().Timestamp().Str("service", "stream_server").Logger()
	pidFile := prog + ".pid"
	pid := os.Getpid()
	logger.Print("Running pid: %d", pid)

	err = ioutil.WriteFile(pidFile, []byte(strconv.Itoa(pid)), 0600)
	if err != nil {
		logger.Print("error writing pid file: " + err.Error())
		os.Exit(1)
	}

	var api API
	if err = api.Init(); err != nil {
		logger.Print("error initializing API server: " + err.Error())
		os.Exit(1)
	}

	// run Kafka consumer manager on a thread
	api.Kafka = KafkaInit()
	go func(k *Kafka) {
		k.Connect()
	}(&api.Kafka)

	// run REST service on a thread
	go func(server *http.Server) {
		logger.Printf("starting api server at %s", server.Addr)
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			logger.Print("error starting API server: " + err.Error())
		}
	}(api.Server)

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	// this will block until context is cancelled or program cancelled
	select {
	case <-sigterm:
		logger.Print("terminating: via signal")
	}

	logger.Print("closing Kafka consumer")
	api.Kafka.cancel()
	if err = api.Kafka.client.Close(); err != nil {
		logger.Print("Error closing client: +", err.Error())
		os.Exit(1)
	}
	logger.Print("closing server")
	api.Server.Close()

	err = os.Remove(pidFile)
	if err != nil {
		logger.Print("error removing pid file: " + err.Error())
		os.Exit(1)
	}

	logger.Print("service stopped")
}
