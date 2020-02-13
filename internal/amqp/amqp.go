package amqp

import (
	"fmt"
	"log"
	"time"

	"github.com/streadway/amqp"
)

// Produce - map of producing channels
var Produce map[string]chan string

// Consume - map of consuming channels
var Consume map[string]chan string

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

func declareProducers(ch *amqp.Channel) error {
	for key := range Produce {
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
		go handleProducer(ch, queue, Produce[key])
	}
	return nil
}

func handleConsumer(queue <-chan amqp.Delivery, goch chan string) {
	for msg := range queue {
		goch <- string(msg.Body)
	}
}

func declareConsumers(ch *amqp.Channel) error {
	for key := range Consume {
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
		go handleConsumer(queue, Consume[key])
	}
	return nil
}

// connect - initializes and starts up the amqp client
func connect(addr string) {
	for {
		time.Sleep(1 * time.Second)

		conn, err := amqp.Dial("amqp://" + addr)
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

		err = declareProducers(ch)
		if err != nil {
			log.Println("[AMQP] Failed to declare producer queues")
			continue
		}

		err = declareConsumers(ch)
		if err != nil {
			log.Println("[AMQP] Failed to declare consumer queues")
			continue
		}

		err = <-notify
	}
}

// Setup - sets up the amqp producers and consumers
func Setup(addr string) {
	Produce = make(map[string]chan string)
	Consume = make(map[string]chan string)
	Produce["ingest_job"] = make(chan string)
	Produce["export_job"] = make(chan string)
	Consume["ingest_status"] = make(chan string)
	Consume["export_status"] = make(chan string)

	go connect(addr)
}
