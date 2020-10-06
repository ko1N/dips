package manager

import (
	"encoding/json"
	"fmt"

	"github.com/ko1N/dips/internal/persistence/messages"
	"github.com/ko1N/dips/pkg/pipeline/tracking"
)

func handleJobMessage() {
	for status := range recvJobMessage {
		msg := tracking.JobMessage{}
		err := json.Unmarshal([]byte(status), &msg)
		if err != nil {
			fmt.Printf("unable to unmarshal job message")
			continue
		}

		messageHandler.Store(msg.JobID, messages.Message{
			Type:    uint(msg.Type),
			Message: msg.Message,
		})
	}
}
