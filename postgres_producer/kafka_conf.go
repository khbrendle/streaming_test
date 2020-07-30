package main

import (
	"log"
	"strings"
	"time"

	"github.com/Shopify/sarama"
)

type KafkaConf struct {
	Brokers    *string        `yaml:"brokers"`
	Version    *string        `yaml:"version"`
	WaitMS     *time.Duration `yaml:"waitMS"`
	saramaConf *sarama.Config
	producer   sarama.AsyncProducer
}

func (k *KafkaConf) Connect() error {
	var err error
	log.Println("initializing new sarama config")
	k.saramaConf = sarama.NewConfig()

	log.Println("setting Kafka config values")
	k.saramaConf.Producer.RequiredAcks = sarama.WaitForLocal             // Only wait for the leader to ack
	k.saramaConf.Producer.Flush.Frequency = *k.WaitMS * time.Millisecond // Flush batches every 500ms

	version, err := sarama.ParseKafkaVersion(*k.Version)
	if err != nil {
		return err
	}
	k.saramaConf.Version = version

	brokerList := strings.Split(*k.Brokers, ",")

	// Async producer
	log.Println("starting async producer")
	k.producer, err = sarama.NewAsyncProducer(brokerList, k.saramaConf)
	if err != nil {
		return err
	}
	return nil
}

func (k *KafkaConf) ListenForErrors() {
	var err error
	for err = range k.producer.Errors() {
		log.Println("Failed to write access log entry:", err)
	}
}
