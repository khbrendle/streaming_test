package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	log.Printf("Running pid: %d", os.Getpid())

	var err error
	var api API
	if err = api.Init(); err != nil {
		panic(err)
	}

	// run Kafka consumer manager on a thread
	api.Kafka = KafkaInit()
	go func(k *Kafka) {
		k.Connect()
	}(&api.Kafka)

	// run REST service on a thread
	go func(server *http.Server) {
		log.Printf("starting api server at %s", server.Addr)
		err := server.ListenAndServe()
		if err != nil {
			panic(err)
		}
	}(api.Server)

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	// this will block until context is cancelled or program cancelled
	select {
	case <-sigterm:
		log.Println("terminating: via signal")
	}

	log.Println("closing Kafka consumer")
	api.Kafka.cancel()
	if err = api.Kafka.client.Close(); err != nil {
		log.Panicf("Error closing client: %v", err)
	}
	log.Println("closing server")
	api.Server.Close()
}
