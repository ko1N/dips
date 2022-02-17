package amqp

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/streadway/amqp"
)

const messageBuffer int = 1000

type Message struct {
	Expiration    string
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
	lock                sync.Mutex
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

// TODO: guarantee thread safety

// RegisterProducer - creates a new producer channel and returns it
func (c *Client) RegisterProducer(name string) chan Message {
	c.lock.Lock()
	defer c.lock.Unlock()

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
	c.lock.Lock()
	defer c.lock.Unlock()

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
	c.lock.Lock()
	defer c.lock.Unlock()

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

func (c *Client) CloseResponseConsumer(name string, correlationId string) {
	c.lock.Lock()
	defer c.lock.Unlock()

	// close channel
	close(c.consumers[name].channels[correlationId])
	delete(c.consumers[name].channels, correlationId)
}

// Run - spawns a client in a new goroutine
func (c *Client) run() {
	go func() {
	outer:
		for {
			time.Sleep(1 * time.Second)

			fmt.Println("[AMQP] trying to connect to " + c.server)
			conn, err := amqp.Dial("amqp://" + c.server)
			if err != nil {
				log.Println("[AMQP] Failed to connect")
				continue
			}
			defer conn.Close()

			notify := conn.NotifyClose(make(chan *amqp.Error, 10))

		inner:
			for {
				select {
				case err = <-notify:
					// clear maps
					c.lock.Lock()
					c.registeredProducers = make(map[string]bool)
					c.registeredConsumers = make(map[string]bool)
					c.lock.Unlock()
					break inner

				default:
					// update consumers and producers
					err = c.declareProducers(conn)
					if err != nil {
						log.Println("[AMQP] Failed to declare producer queues: " + err.Error())
						continue outer
					}

					err = c.declareConsumers(conn)
					if err != nil {
						log.Println("[AMQP] Failed to declare consumer queues: " + err.Error())
						continue outer
					}

					time.Sleep(1 * time.Millisecond)
					break
				}
			}
		}
	}()
}

func handleProducer(mqchn *amqp.Channel, q amqp.Queue, chn chan Message) {
	defer mqchn.Close()

	for msg := range chn {
		err := mqchn.Publish("",
			q.Name,
			false,
			false,
			amqp.Publishing{
				ContentType:   "application/json",
				Body:          []byte(msg.Payload),
				CorrelationId: msg.CorrelationId,
				Expiration:    msg.Expiration,
			})
		if err != nil {
			fmt.Printf("[AMQP] Error sending message, requeueing\n")
			chn <- msg // re-queue failed message
			return     // abort goroutine
		}
	}
}

func (c *Client) declareProducers(conn *amqp.Connection) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	for name := range c.producers {
		if !c.registeredProducers[name] {
			fmt.Printf("[AMQP] Creating producer channel %s\n", name)
			mqchn, err := conn.Channel()
			if err != nil {
				log.Println("[AMQP] Failed to open producer channel" + err.Error())
				return err
			}

			fmt.Printf("[AMQP] Creating producer queue %s\n", name)
			queue, err := mqchn.QueueDeclare(
				name,
				true,
				false,
				false,
				false,
				nil)
			if err != nil {
				mqchn.Close()
				return err
			}
			c.registeredProducers[name] = true
			chn := c.producers[name].channels[""]
			go handleProducer(mqchn, queue, chn)
		}
	}
	return nil
}

func (c *Client) handleConsumer(mqchn *amqp.Channel, amqpDelivery <-chan amqp.Delivery, queue *Queue) {
	defer mqchn.Close()

	for msg := range amqpDelivery {
		c.lock.Lock()
		chn := queue.channels[msg.CorrelationId]
		c.lock.Unlock()

		if chn != nil {
			chn <- Message{
				Payload: string(msg.Body),
			}
			msg.Ack(false)
		} else {
			msg.Nack(false, true)
		}
	}
}

func (c *Client) declareConsumers(conn *amqp.Connection) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	for name := range c.consumers {
		if !c.registeredConsumers[name] {
			fmt.Printf("[AMQP] Creating consumer channel %s\n", name)
			mqchn, err := conn.Channel()
			if err != nil {
				log.Println("[AMQP] Failed to open consumer channel" + err.Error())
				return err
			}

			fmt.Printf("[AMQP] Creating consumer queue %s\n", name)
			_, err = mqchn.QueueDeclare(
				name,
				true,
				false,
				false,
				false,
				nil)
			if err != nil {
				mqchn.Close()
				return err
			}
			queue, err := mqchn.Consume(
				name,
				"",
				false, // autoAck
				false,
				false,
				false,
				nil)
			if err != nil {
				mqchn.Close()
				return err
			}
			c.registeredConsumers[name] = true
			go c.handleConsumer(mqchn, queue, c.consumers[name])
		}
	}
	return nil
}
