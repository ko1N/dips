package manager

import (
	"github.com/ko1N/dips/internal/persistence/messages"
	"github.com/ko1N/dips/pkg/client"
)

func handleJobMessages(ev *client.EventHandler) {
	ev.HandleMessage(func(msg *client.MessageEvent) error {
		messageHandler.Store(msg.JobID, messages.Message{
			Type:    uint(msg.Type),
			Message: msg.Message,
		})
		return nil
	})
}
