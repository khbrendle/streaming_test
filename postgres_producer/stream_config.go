package main

import (
	"time"

	"github.com/lib/pq"
)

type StreamConfig struct {
	Postgres PostgresListenerConfig `yaml:"postgres"`
	// Kafka topic to send notification to
	Kafka KafkaProducerConfig `yaml:"kafka"`
}

type PostgresListenerConfig struct {
	// Channel is the database channel name i.e. from the `LISTEN {channel}` SQL statement
	Channel *string `yaml:"channel"`
	// Reconnect params for PG side
	MinReconnectInterval time.Duration `yaml:"minReconnectInterval"`
	MaxReconnectInterval time.Duration `yaml:"maxReconnectInterval"`
	listener             *pq.Listener
}

type KafkaProducerConfig struct {
	Topic *string `yaml:"topic"`
	Key   *string `yaml:"key"`
}
