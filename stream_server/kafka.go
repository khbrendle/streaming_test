package main

import (
	"context"
	"sync"

	"github.com/Shopify/sarama"
)

type Kafka struct {
	Brokers  []string
	Version  string
	Group    string
	Assignor string
	Oldest   bool
	Config   *sarama.Config
	ready    chan bool
	// topics & subscriptions are kept private to enforce adding/removing topics via
	// methods to keep both in sync
	stLock sync.RWMutex
	// list of topics to receive messages for
	topics *[]string
	// map of channels for each topic to receive messages on
	subs map[string]*MessageSub
	// context controls closing the Kafka connection
	ctx    context.Context
	cancel func()
	client sarama.ConsumerGroup
}

type MessageSub struct {
	counter *uint32
	// messages *chan *sarama.ConsumerMessage
	// each connected client has a channel which gets the message
	clients map[string]*chan *sarama.ConsumerMessage
}

func KafkaInit() (k Kafka) {
	var err error
	topicsArray := []string{"customer_count", "order_count"}
	k = Kafka{
		Brokers: []string{"127.0.0.1:9092"},
		Version: "2.5.0",
		Group:   "example",
		// topics:   []string{"test"},
		topics:   &topicsArray,
		Assignor: "roundrobin",
		Oldest:   false,
	}

	// this contains subscriptions
	// key is Kafka topic, value contains counter and map of clients
	k.subs = make(map[string]*MessageSub)

	for _, topic := range topicsArray {

		logger.Print("initializing topic subscription information for topic " + topic)
		k.subs[topic] = &MessageSub{
			// initialize connection count
			counter: Uint32(1),
			// initialize message channel
			clients: make(map[string]*chan *sarama.ConsumerMessage),
		}
	}

	k.ready = make(chan bool)
	k.ctx, k.cancel = context.WithCancel(context.TODO())

	k.Config = sarama.NewConfig()

	version, err := sarama.ParseKafkaVersion(k.Version)
	if err != nil {
		logger.Error().Msg(err.Error())
	}
	k.Config.Version = version

	switch k.Assignor {
	case "sticky":
		k.Config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategySticky
	case "roundrobin":
		k.Config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	case "range":
		k.Config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRange
	default:
		logger.Warn().Msgf("Unrecognized consumer group partition assignor: %s", k.Assignor)
	}

	if k.Oldest {
		k.Config.Consumer.Offsets.Initial = sarama.OffsetOldest
	}
	return
}

func (k *Kafka) Connect() {
	var err error

	k.client, err = sarama.NewConsumerGroup(k.Brokers, k.Group, k.Config)
	if err != nil {
		logger.Error().Msgf("Error creating consumer group client: %v", err)
	}

	go func(kaf *Kafka) {
		for {
			// `Consume` should be called inside an infinite loop, when a
			// server-side rebalance happens, the consumer session will need to be
			// recreated to get the new claims

			// log.Println("setting read lock of kafka info")
			// k.stLock.RLock()
			// if len(*kaf.topics) == 0 {
			// 	// log.Println("no topics")
			// 	// log.Println("unsetting read lock of kafka info")
			// 	// k.stLock.RUnlock()
			// 	continue
			// }
			logger.Printf("consuming messages for topics %v\n", *kaf.topics)
			if err := kaf.client.Consume(kaf.ctx, *kaf.topics, kaf); err != nil {
				// free the lock on error
				// log.Println("unsetting read lock of kafka info")
				// k.stLock.RUnlock()

				logger.Printf("Error from consumer: %v", err)
			}

			// log.Println("unsetting read lock of kafka info")
			// k.stLock.RUnlock()
			// check if context was cancelled, signaling that the consumer should stop
			if kaf.ctx.Err() != nil {
				return
			}
			kaf.ready = make(chan bool)
		}
	}(k)

	<-k.ready // Await till the consumer has been set up
	logger.Print("Sarama consumer up and running!...")

	// context cancel func to stop Kafka connection
	<-k.ctx.Done()
}

// this is not locking so needs to happen between lock/unlock
func (k *Kafka) TopicSubscribed(topic *string) (exists bool) {
	_, exists = k.subs[*topic]
	return
}

func (k *Kafka) Subscribe(clientID *string, topic *string) {
	// update topics slice
	logger.Print("locking kafka metadata")
	k.stLock.Lock()
	// Check if the topic has been initialized
	// if k.TopicSubscribed(topic) { // TODO: using individual consumers per topic would use this
	logger.Print("topic subscribed")
	// if yes then increment counter
	logger.Print("getting topic counter")
	c := k.subs[*topic].counter

	logger.Print("setting incrementd topic counter")
	k.subs[*topic].counter = Uint32(*c + 1)

	// TODO: using individual consumers per topic would use this
	// } else {
	// 	log.Println("topic not yet subscribed")
	// 	// if topic not initialized, then add to list, start counter, & create channel
	// 	log.Println("addding topic to consumer list")
	// 	*k.topics = append(*k.topics, *topic)
	//
	// 	log.Println("initializing topic subscription information")
	// 	k.subs[*topic] = &MessageSub{
	// 		// initialize connection count
	// 		counter: Uint32(1),
	// 		// initialize message channel
	// 		clients: make(map[string]*chan *sarama.ConsumerMessage),
	// 	}
	// }

	logger.Print("initializing client channel")
	k.subs[*topic].clients[*clientID] = NewChannel()

	logger.Print("unlocking kafka metadata")
	k.stLock.Unlock()

	return
}

// Removing topics requires careful usage of locks
func (k *Kafka) Unsubscribe(clientID *string, topic *string) (err error) {
	// var t []string

	logger.Print("locking kafka metadata")
	k.stLock.Lock()

	logger.Print("getting topic counter")
	c := k.subs[*topic].counter
	// if removing last subscriptoin, then stop consuming from Kafka

	// TODO: using individual consumers per topic would use this
	// if *c == 1 {
	// 	log.Println("last topic subscriber, removing Kafka consumer")
	//
	// 	// remove topic from slice
	// 	if t, err = SliceRemoveString(*k.topics, *topic); err != nil {
	// 		log.Println("unlocking kafka metadata")
	// 		k.stLock.Unlock()
	//
	// 		return err
	// 	}
	//
	// 	log.Println("setting updated topics list")
	// 	*k.topics = t
	//
	// 	// delete subscription
	// 	log.Println("deleting topic subscription")
	// 	delete(k.subs, *topic)
	//
	// 	log.Println("unlocking kafka metadata")
	// 	k.stLock.Unlock()
	//
	// 	return nil
	// }

	// otherwise decrement the counter
	logger.Print("setting updated counter")
	k.subs[*topic].counter = Uint32(*c - 1)

	logger.Print("removing client channel")
	delete(k.subs[*topic].clients, *clientID)

	logger.Print("unlocking kafka metadata")
	k.stLock.Unlock()

	return nil
}

func (k *Kafka) GetMessage(clientID *string, topic *string) *chan *sarama.ConsumerMessage {
	return k.subs[*topic].clients[*clientID]
}

func NewChannel() *chan *sarama.ConsumerMessage {
	c := make(chan *sarama.ConsumerMessage, 5)
	return &c
}

/*
  these next 3 methods are to satify the ConsumerGroupHanlder interface
*/

// Setup is run at the beginning of a new session, before ConsumeClaim
func (k *Kafka) Setup(sarama.ConsumerGroupSession) error {
	// Mark the consumer as ready
	close(k.ready)
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (k *Kafka) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (k *Kafka) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	// NOTE:
	// Do not move the code below to a goroutine.
	// The `ConsumeClaim` itself is called within a goroutine, see:
	// https://github.com/Shopify/sarama/blob/master/consumer_group.go#L27-L29
	var (
		topic *MessageSub
		ok    bool
	)
	for message := range claim.Messages() {
		logger.Printf("Message claimed: value = %s, timestamp = %v, topic = %s", string(message.Value), message.Timestamp, message.Topic)
		session.MarkMessage(message, "")
		// TODO: evaluate necessity of locks here
		// there will only be single thread per topic so it should not need to lock

		// log.Println("locing kafka metadata")
		// k.stLock.Lock()

		// log.Println("sending topic messges to clients")
		topic, ok = k.subs[message.Topic]
		if !ok {
			// log.Printf("warn!!! no subscriptions for topic %s", message.Topic)
			continue
		}
		for _, ch := range topic.clients {
			*ch <- message
		}
		// *k.subs[message.Topic].messages <- message

		// log.Println("unlocing kafka metadata")
		// k.stLock.Unlock()
	}

	return nil
}
