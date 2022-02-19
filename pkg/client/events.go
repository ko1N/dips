package client

import (
	"encoding/json"

	"github.com/ko1N/dips/internal/amqp"
)

// The event to be dispatched
type Event struct {
	client   *Client
	status   *StatusEvent
	message  *MessageEvent
	variable *VariableEvent
}

// the type of the job status update
type StatusEventType uint

const (
	// progress update
	ProgressEvent StatusEventType = 1
)

type StatusEvent struct {
	JobId    string
	TaskIdx  uint
	Type     StatusEventType
	Progress uint
	//JobStatus string // TODO: enum
}

// the type of the message
type MessageEventType uint

const (
	StatusMessage MessageEventType = 0
	ErrorMessage  MessageEventType = 1
	StdInMessage  MessageEventType = 2
	StdOutMessage MessageEventType = 3
	StdErrMessage MessageEventType = 4
)

type MessageEvent struct {
	JobId   string
	TaskIdx uint
	Type    MessageEventType
	Message string
}

type VariableEvent struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
}

func (c *Client) NewEvent() *Event {
	return &Event{
		client: c,
	}
}

func (e *Event) Status(status *StatusEvent) *Event {
	e.status = status
	return e
}

func (e *Event) Message(message *MessageEvent) *Event {
	e.message = message
	return e
}

func (e *Event) Variable(variable *VariableEvent) *Event {
	e.variable = variable
	return e
}

// Dispatches the event (and never blocks)
func (e *Event) Dispatch() {
	if e.status != nil {
		queue := e.client.amqp.RegisterProducer("dips.event.status")

		request, err := json.Marshal(&e.status)
		if err != nil {
			panic("Invalid status event: " + err.Error())
		}

		queue <- amqp.Message{
			Payload: string(request),
		}
	}

	if e.message != nil {
		queue := e.client.amqp.RegisterProducer("dips.event.message")

		request, err := json.Marshal(&e.message)
		if err != nil {
			panic("Invalid message event: " + err.Error())
		}

		queue <- amqp.Message{
			Payload: string(request),
		}
	}

	if e.variable != nil {
		queue := e.client.amqp.RegisterProducer("dips.event.variable")

		request, err := json.Marshal(&e.variable)
		if err != nil {
			panic("Invalid variable event: " + err.Error())
		}

		queue <- amqp.Message{
			Payload: string(request),
		}
	}
}

type EventHandler struct {
	client          *Client
	statusHandler   func(*StatusEvent) error
	messageHandler  func(*MessageEvent) error
	variableHandler func(*VariableEvent) error
}

func (c *Client) NewEventHandler() *EventHandler {
	return &EventHandler{
		client: c,
	}
}

func (h *EventHandler) HandleStatus(status func(*StatusEvent) error) *EventHandler {
	h.statusHandler = status
	return h
}

func (h *EventHandler) HandleMessage(message func(*MessageEvent) error) *EventHandler {
	h.messageHandler = message
	return h
}

func (h *EventHandler) HandleVariable(variable func(*VariableEvent) error) *EventHandler {
	h.variableHandler = variable
	return h
}

// Run - Starts a new goroutine for this event handler
func (h *EventHandler) Run() {
	// TODO: graceful shutdown
	if h.statusHandler != nil {
		go func() {
			queue := h.client.amqp.RegisterProducer("dips.event.status")
			for request := range queue {
				var statusEvent StatusEvent
				err := json.Unmarshal([]byte(request.Payload), &statusEvent)
				if err != nil {
					panic("Invalid status event: " + err.Error())
				}
				h.statusHandler(&statusEvent)
			}
		}()
	}

	if h.messageHandler != nil {
		go func() {
			queue := h.client.amqp.RegisterProducer("dips.event.message")
			for request := range queue {
				var messageEvent MessageEvent
				err := json.Unmarshal([]byte(request.Payload), &messageEvent)
				if err != nil {
					panic("Invalid log event: " + err.Error())
				}
				h.messageHandler(&messageEvent)
			}
		}()
	}

	if h.variableHandler != nil {
		go func() {
			queue := h.client.amqp.RegisterProducer("dips.event.variable")
			for request := range queue {
				var variableEvent VariableEvent
				err := json.Unmarshal([]byte(request.Payload), &variableEvent)
				if err != nil {
					panic("Invalid variable event: " + err.Error())
				}
				h.variableHandler(&variableEvent)
			}
		}()
	}
}
