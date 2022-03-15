package dipscl

import "github.com/ko1N/dips/internal/amqp"

// Client - Dips client instance
type Client struct {
	amqp        *amqp.Client
	statusQueue (chan amqp.Message)
	logQueue    (chan amqp.Message)
}

// NewClient - Creates a new Dips client
func NewClient(host string) (*Client, error) {
	amqp := amqp.NewAMQP(amqp.Config{
		Host: host,
	})
	return &Client{
		amqp:        amqp,
		statusQueue: amqp.RegisterProducer("dips.worker.status"),
		logQueue:    amqp.RegisterProducer("dips.worker.log"),
	}, nil
}
