package client

import "github.com/ko1N/dips/internal/amqp"

type Client struct {
	amqp        *amqp.Client
	workers     []Worker
	updateQueue (chan string)
}

// TODO: custom client config
func NewClient(host string) (*Client, error) {
	client := amqp.Create(amqp.Config{
		Host: host,
	})

	/*
		recvPipelineExecute := client.RegisterConsumer("pipeline_execute")
		sendJobStatus := client.RegisterProducer("job_status")
		sendJobMessage := client.RegisterProducer("job_message")
	*/

	return &Client{
		amqp:    client,
		workers: []Worker{},
	}, nil
}

func (self *Client) Run() {
	self.amqp.Run()
}
