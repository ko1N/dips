package manager

import (
	"encoding/json"
	"fmt"

	"gitlab.strictlypaste.xyz/ko1n/dips/internal/persistence/messages"
	"gitlab.strictlypaste.xyz/ko1n/dips/pkg/pipeline"
)

func handleJobMessage() {
	for status := range recvJobMessage {
		msg := pipeline.JobMessage{}
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
