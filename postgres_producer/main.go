package main

import (
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/Shopify/sarama"
	"github.com/lib/pq"
	"github.com/rs/zerolog"
	yaml "gopkg.in/yaml.v1"
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

	logger = zerolog.New(logFile).With().Timestamp().Str("service", "postgres_producer").Logger()
	pidFile := prog + ".pid"

	if os.Args[1] == "" {
		logger.Print("single argument config file path")
		os.Exit(1)
	}

	pid := os.Getpid()
	logger.Printf("running with pid %d", pid)

	err = ioutil.WriteFile(pidFile, []byte(strconv.Itoa(pid)), 0600)
	if err != nil {
		logger.Print("error writing pid file: " + err.Error())
		os.Exit(1)
	}

	var b []byte
	logger.Print("reading config file")
	if b, err = ioutil.ReadFile(os.Args[1]); err != nil {
		logger.Print("error reading config file: " + err.Error())
	}

	var conf StartPGListenerInput
	logger.Print("parsing config file")
	if err = yaml.Unmarshal(b, &conf); err != nil {
		logger.Print("error parsing config YAML: " + err.Error())
	}

	logger.Print("initializing Kafka connection")
	if err = conf.KafkaConf.Connect(); err != nil {
		logger.Print("error connecting Kafka: " + err.Error())
	}

	// We will just log to STDOUT if we're not able to produce messages.
	// Note: messages will only be returned here after all retry attempts are exhausted.
	logger.Print("starting Kafka error listener")
	go conf.KafkaConf.ListenForErrors()

	// these errors should not be fatal so just printing
	conf.listenFromDB()

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	// this will block until context is cancelled or program cancelled
	select {
	case <-sigterm:
		logger.Print("got signal to cancel")
		if err = conf.KafkaConf.producer.Close(); err != nil {
			logger.Print("erro closing Kafka connection: " + err.Error())
		}
		for _, stream := range conf.Streams {
			if err = stream.Postgres.listener.Close(); err != nil {
				logger.Printf("error closing stream for channel `%s`: %e", *stream.Postgres.Channel, err)
			}
		}
	}

	// clean up pid file
	err = os.Remove(pidFile)
	if err != nil {
		logger.Print("error deleting pid file: " + err.Error())
		os.Exit(1)
	}

}

type StartPGListenerInput struct {
	PostgresConf `yaml:"postgresConf"`
	KafkaConf    `yaml:"kafkaConf"`
	Streams      []*StreamConfig `yaml:"streams"`
}

func (c *StartPGListenerInput) listenFromDB() {
	var err error

	for _, s := range c.Streams {

		// initialize listener
		logger.Print("initilizing Postgres listener for channel " + *s.Postgres.Channel)
		s.Postgres.listener = pq.NewListener(c.PostgresConf.ConnectionString(), s.Postgres.MinReconnectInterval, s.Postgres.MaxReconnectInterval, func(ev pq.ListenerEventType, err error) {
			if err != nil {
				logger.Printf("Failed to start listener: %s", err)
				os.Exit(1)
			}
		})

		// start listener
		logger.Print("starting listener")
		if err = s.Postgres.listener.Listen(*s.Postgres.Channel); err != nil {
			// if error then print message about failure
			logger.Printf("listener failed to establish for channel '%s'", *s.Postgres.Channel)
			// skip rest of listener initialization
			continue
		}

		logger.Printf(
			"Listening to notifications on channel \"%s\"",
			*s.Postgres.Channel)

		// start listener in goroutine to process each stream
		go func(c *StreamConfig, o *sarama.AsyncProducer) {
			for {
				select {
				case n := <-c.Postgres.listener.Notify:
					if n != nil {
						(*o).Input() <- &sarama.ProducerMessage{
							Topic: *c.Kafka.Topic,
							Key:   sarama.StringEncoder(*c.Kafka.Key),
							Value: sarama.StringEncoder(n.Extra),
						}
					}
				}
			}
		}(s, &c.KafkaConf.producer)
	}

	logger.Print("Start processing notifications, waiting for events…")
}

func String(s string) *string {
	return &s
}
