package amqp

import (
	"fmt"
	"log"
	"time"

	"github.com/streadway/amqp"
)

// Client - Simple AMQP Client wrapper
type Client struct {
	server    string
	producers map[string]chan string
	consumers map[string]chan string
	signal    chan struct{}
}

// Create - will create a new AMQP Client object
func Create(server string) Client {
	return Client{
		server:    server,
		producers: make(map[string]chan string),
		consumers: make(map[string]chan string),
	}
}

// RegisterProducer - creates a new producer channel and returns it
func (c *Client) RegisterProducer(name string) chan string {
	chn := make(chan string)
	c.producers[name] = chn
	return chn
}

// RegisterConsumer - creates a new consumer channel and returns it
func (c *Client) RegisterConsumer(name string) chan string {
	chn := make(chan string)
	c.consumers[name] = chn
	return chn
}

// Start - spawns a client in a new goroutine
func (c *Client) Start() {
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
	for {
		select {
		case body := <-goch:
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
}

func (c *Client) declareProducers(ch *amqp.Channel) error {
	for key := range c.producers {
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
