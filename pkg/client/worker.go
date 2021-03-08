package client

type Worker struct {
	serviceQueue (chan string)
}

// TODO: sanitize name
func (self *Client) NewWorker(name string) *Worker {
	serviceQueue := self.amqp.RegisterConsumer("worker_" + name)
	return &Worker{
		serviceQueue: serviceQueue,
	}
}

func (self *Worker) Handler(func(*Job) error) *Worker {
	// TODO:
	return self
}
