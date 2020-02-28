package amqp

import (
	"encoding/json"

	log "github.com/inconshreveable/log15"
)

// logHandler - Logging handler which sends data over amqp
type logHandler struct {
	ID      string
	Channel chan string
}

// logMessage - Describes a log message sent over amqp
type logMessage struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

// CreateLogHandler - Creates a new AMQP logHandler
func LogHandler(id string, channel chan string) log.Handler {
	return &logHandler{
		ID:      id,
		Channel: channel,
	}
}

// Log - sends the log message over amqp
func (h *logHandler) Log(r *log.Record) error {
	msg := logMessage{
		ID:      h.ID,
		Message: r.Msg,
	}
	str, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	h.Channel <- string(str)
	return nil
}
