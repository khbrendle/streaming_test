package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Shopify/sarama"
	"github.com/lib/pq"
	yaml "gopkg.in/yaml.v1"
)

func main() {

	if os.Args[1] == "" {
		log.Fatal("single argument config file path")
	}

	fmt.Printf("running with pid %d\n", os.Getpid())

	var err error
	var b []byte
	log.Println("reading config file")
	if b, err = ioutil.ReadFile(os.Args[1]); err != nil {
		log.Fatal("error reading config file: " + err.Error())
	}

	var conf StartPGListenerInput
	log.Println("parsing config file")
	if err = yaml.Unmarshal(b, &conf); err != nil {
		log.Fatal("error parsing config YAML: " + err.Error())
	}

	log.Println("initializing Kafka connection")
	if err = conf.KafkaConf.Connect(); err != nil {
		log.Fatal("error connecting Kafka: " + err.Error())
	}

	// We will just log to STDOUT if we're not able to produce messages.
	// Note: messages will only be returned here after all retry attempts are exhausted.
	log.Println("starting Kafka error listener")
	go conf.KafkaConf.ListenForErrors()

	conf.listenFromDB()

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	// this will block until context is cancelled or program cancelled
	select {
	case <-sigterm:
		if err = conf.KafkaConf.producer.Close(); err != nil {
			log.Println("erro closing Kafka connection: " + err.Error())
		}
		for _, stream := range conf.Streams {
			if err = stream.Postgres.listener.Close(); err != nil {
				log.Printf("error closing stream for channel `%s`: %e", *stream.Postgres.Channel, err)
			}
		}
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
		log.Println("initilizing Postgres listener for channel " + *s.Postgres.Channel)
		s.Postgres.listener = pq.NewListener(c.PostgresConf.ConnectionString(), s.Postgres.MinReconnectInterval, s.Postgres.MaxReconnectInterval, func(ev pq.ListenerEventType, err error) {
			if err != nil {
				log.Printf("Failed to start listener: %s", err)
				os.Exit(1)
			}
		})

		// start listener
		log.Println("starting listener")
		if err = s.Postgres.listener.Listen(*s.Postgres.Channel); err != nil {
			// if error then print message about failure
			log.Printf("listener failed to establish for channel '%s'", *s.Postgres.Channel)
			// skip rest of listener initialization
			continue
		}

		log.Printf(
			"Listening to notifications on channel \"%s\"",
			*s.Postgres.Channel)

		// start listener in goroutine to process each stream
		go func(c *StreamConfig, o *sarama.AsyncProducer) {
			for {
				select {
				case n := <-c.Postgres.listener.Notify:
					(*o).Input() <- &sarama.ProducerMessage{
						Topic: *c.Kafka.Topic,
						Key:   sarama.StringEncoder(*c.Kafka.Key),
						Value: sarama.StringEncoder(n.Extra),
					}
				}
			}
		}(s, &c.KafkaConf.producer)
	}

	log.Println("Start processing notifications, waiting for events…")
}

func String(s string) *string {
	return &s
}
