package amqp

import (
	"fmt"
	"log"
	"time"

	"github.com/streadway/amqp"
)

const messageBuffer int = 1000

// Client - Simple AMQP Client wrapper
type Client struct {
	server    string
	producers map[string]chan string
	consumers map[string]chan string
	signal    chan struct{}
}

// Config - config entry describing a amqp config
type Config struct {
	Host string `json:"host" toml:"host"`
}

// NewAMQP - will create a new AMQP Client object
func NewAMQP(conf Config) *Client {
	client := Client{
		server:    conf.Host,
		producers: make(map[string]chan string),
		consumers: make(map[string]chan string),
	}
	client.run()
	return &client
}

// TODO: allow producers/consumers to be registered while amqp is running :)
// TODO: guarantee thread safety

// RegisterProducer - creates a new producer channel and returns it
func (c *Client) RegisterProducer(name string) chan string {
	if c.producers[name] != nil {
		return c.producers[name]
	}
	chn := make(chan string, messageBuffer)
	c.producers[name] = chn
	return chn
}

// RegisterConsumer - creates a new consumer channel and returns it
func (c *Client) RegisterConsumer(name string) chan string {
	if c.consumers[name] != nil {
		return c.consumers[name]
	}
	chn := make(chan string, messageBuffer)
	c.consumers[name] = chn
	return chn
}

// Run - spawns a client in a new goroutine
func (c *Client) run() {
	go func() {
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

			ch, err := conn.Channel()
			if err != nil {
				log.Println("[AMQP] Failed to open channel")
				continue
			}
			defer ch.Close()

			err = c.declareProducers(ch)
			if err != nil {
				log.Println("[AMQP] Failed to declare producer queues")
				continue
			}

			err = c.declareConsumers(ch)
			if err != nil {
				log.Println("[AMQP] Failed to declare consumer queues")
				continue
			}

			// TODO: lazily create producers/consumers here

			err = <-notify
		}
	}()
}

func handleProducer(ch *amqp.Channel, q amqp.Queue, goch chan string) {
	for body := range goch {
		err := ch.Publish("",
			q.Name,
			false,
			false,
			amqp.Publishing{
				ContentType: "application/json",
				Body:        []byte(body),
			})
		if err != nil {
			fmt.Printf("[AMQP] Error sending message\n")
			goch <- body // re-queue failed message
			return       // abort goroutine
		}
	}
}

func (c *Client) declareProducers(ch *amqp.Channel) error {
	for key := range c.producers {
		fmt.Printf("[AMQP] Creating queue %s\n", key)
		queue, err := ch.QueueDeclare(
			key,
			true,
			false,
			false,
			false,
			nil)
		if err != nil {
			return err
		}
		go handleProducer(ch, queue, c.producers[key])
	}
	return nil
}

func handleConsumer(queue <-chan amqp.Delivery, goch chan string) {
	for msg := range queue {
		goch <- string(msg.Body)
	}
}

func (c *Client) declareConsumers(ch *amqp.Channel) error {
	for key := range c.consumers {
		fmt.Printf("[AMQP] Creating queue %s\n", key)
		_, err := ch.QueueDeclare(
			key,
			true,
			false,
			false,
			false,
			nil)
		if err != nil {
			return err
		}
		queue, err := ch.Consume(
			key,
			"",
			true,
			false,
			false,
			false,
			nil)
		go handleConsumer(queue, c.consumers[key])
	}
	return nil
}
