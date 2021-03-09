package client

import "encoding/json"

type Event struct {
	client   *Client
	status   *StatusEvent
	log      *LogEvent
	variable *VariableEvent
}

type StatusEvent struct {
}

type LogEvent struct {
}

type VariableEvent struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
}

func (client *Client) NewEvent() *Event {
	return &Event{
		client: client,
	}
}

func (event *Event) Status(status *StatusEvent) *Event {
	event.status = status
	return event
}

func (event *Event) Log(log *LogEvent) *Event {
	event.log = log
	return event
}

func (event *Event) Variable(variable *VariableEvent) *Event {
	event.variable = variable
	return event
}

func (event *Event) Dispatch() {
	if event.status != nil {
		queue := event.client.amqp.RegisterProducer("dips.event.status")

		request, err := json.Marshal(&event.status)
		if err != nil {
			panic("Invalid status event: " + err.Error())
		}

		queue <- string(request)
	}

	if event.log != nil {
		queue := event.client.amqp.RegisterProducer("dips.event.log")

		request, err := json.Marshal(&event.log)
		if err != nil {
			panic("Invalid log event: " + err.Error())
		}

		queue <- string(request)
	}

	if event.variable != nil {
		queue := event.client.amqp.RegisterProducer("dips.event.variable")

		request, err := json.Marshal(&event.variable)
		if err != nil {
			panic("Invalid variable event: " + err.Error())
		}

		queue <- string(request)
	}
}
