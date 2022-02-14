package amqp

import (
	"log"
	"time"

	"github.com/streadway/amqp"
)

const messageBuffer int = 1000

type Message struct {
	CorrelationId string
	Payload       string
}

type Queue struct {
	// Mapping from correlation id to go channel
	// By default this maps an empty correlation id to the only channel
	channels map[string]chan Message
}

// Client - Simple AMQP Client wrapper
type Client struct {
	server              string
	producers           map[string]*Queue
	consumers           map[string]*Queue
	registeredProducers map[string]bool
	registeredConsumers map[string]bool
}

// Config - config entry describing a amqp config
type Config struct {
	Host string `json:"host" toml:"host"`
}

// NewAMQP - will create a new AMQP Client object
func NewAMQP(conf Config) *Client {
	client := Client{
		server:              conf.Host,
		producers:           make(map[string]*Queue),
		consumers:           make(map[string]*Queue),
		registeredProducers: make(map[string]bool),
		registeredConsumers: make(map[string]bool),
	}
	client.run()
	return &client
}

// TODO: allow producers/consumers to be registered while amqp is running :)
// TODO: guarantee thread safety

// RegisterProducer - creates a new producer channel and returns it
func (c *Client) RegisterProducer(name string) chan Message {
	if c.producers[name] != nil {
		return c.producers[name].channels[""]
	}
	chn := make(chan Message, messageBuffer)
	c.producers[name] = &Queue{
		channels: map[string]chan Message{"": chn},
	}
	return chn
}

// RegisterConsumer - creates a new consumer channel and returns it
func (c *Client) RegisterConsumer(name string) chan Message {
	if c.consumers[name] != nil {
		if c.consumers[name].channels[""] != nil {
			return c.consumers[name].channels[""]
		} else {
			chn := make(chan Message, messageBuffer)
			c.consumers[name].channels[""] = chn
			return chn
		}
	}
	chn := make(chan Message, messageBuffer)
	c.consumers[name] = &Queue{
		channels: map[string]chan Message{"": chn},
	}
	return chn
}

// RegisterConsumer - creates a new consumer channel and returns it
func (c *Client) RegisterResponseConsumer(name string, correlationId string) chan Message {
	if c.consumers[name] != nil {
		if c.consumers[name].channels[correlationId] != nil {
			return c.consumers[name].channels[correlationId]
		} else {
			chn := make(chan Message, messageBuffer)
			c.consumers[name].channels[correlationId] = chn
			return chn
		}
	}
	chn := make(chan Message, messageBuffer)
	c.consumers[name] = &Queue{
		channels: map[string]chan Message{correlationId: chn},
	}
	return chn
}

// Run - spawns a client in a new goroutine
func (c *Client) run() {
	go func() {
	outer:
		for {
			time.Sleep(1 * time.Second)

			//fmt.Println("[AMQP] trying to connect to " + c.server)
			conn, err := amqp.Dial("amqp://" + c.server)
			if err != nil {
				log.Println("[AMQP] Failed to connect")
				continue
			}
			defer conn.Close()

			notify := conn.NotifyClose(make(chan *amqp.Error, 10))

			ch, err := conn.Channel()
			if err != nil {
				log.Println("[AMQP] Failed to open channel")
				continue
			}
			defer ch.Close()

			// TODO: do not recreate producers/consumers

		inner:
			for {
				select {
				case err = <-notify:
					// clear maps
					c.registeredProducers = make(map[string]bool)
					c.registeredConsumers = make(map[string]bool)
					break inner

				default:
					// update consumers and producers
					err = c.declareProducers(ch)
					if err != nil {
						log.Println("[AMQP] Failed to declare producer queues")
						continue outer
					}

					err = c.declareConsumers(ch)
					if err != nil {
						log.Println("[AMQP] Failed to declare consumer queues")
						continue outer
					}

					time.Sleep(1 * time.Millisecond)
					break
				}
			}
		}
	}()
}

func handleProducer(amqpChannel *amqp.Channel, q amqp.Queue, queue *Queue) {
	for msg := range queue.channels[""] {
		//fmt.Println("producer correlationId: " + msg.CorrelationId)
		err := amqpChannel.Publish("",
			q.Name,
			false,
			false,
			amqp.Publishing{
				ContentType:   "application/json",
				Body:          []byte(msg.Payload),
				CorrelationId: msg.CorrelationId,
			})
		if err != nil {
			//fmt.Printf("[AMQP] Error sending message, requeueing\n")
			queue.channels[""] <- msg // re-queue failed message
			return                    // abort goroutine
		}
	}
}

func (c *Client) declareProducers(ch *amqp.Channel) error {
	for name := range c.producers {
		if !c.registeredProducers[name] {
			//fmt.Printf("[AMQP] Creating producer queue %s\n", name)
			queue, err := ch.QueueDeclare(
				name,
				true,
				false,
				false,
				false,
				nil)
			if err != nil {
				return err
			}
			c.registeredProducers[name] = true
			go handleProducer(ch, queue, c.producers[name])
		}
	}
	return nil
}

func handleConsumer(amqpDelivery <-chan amqp.Delivery, queue *Queue) {
	for msg := range amqpDelivery {
		//fmt.Println("consumer correlationId: " + msg.CorrelationId)
		if queue.channels[msg.CorrelationId] != nil {
			queue.channels[msg.CorrelationId] <- Message{
				Payload: string(msg.Body),
			}
			msg.Ack(false)
		} else {
			msg.Nack(false, true)
		}
	}
}

func (c *Client) declareConsumers(ch *amqp.Channel) error {
	for name := range c.consumers {
		if !c.registeredConsumers[name] {
			//fmt.Printf("[AMQP] Creating consumer queue %s\n", name)
			_, err := ch.QueueDeclare(
				name,
				true,
				false,
				false,
				false,
				nil)
			if err != nil {
				return err
			}
			queue, err := ch.Consume(
				name,
				"",
				false, // autoAck
				false,
				false,
				false,
				nil)
			c.registeredConsumers[name] = true
			go handleConsumer(queue, c.consumers[name])
		}
	}
	return nil
}
