package messages

import (
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/influxdata/influxdb1-client" // this is important because of the bug in go mod
	client "github.com/influxdata/influxdb1-client/v2"
)

// MessageHandler - wrapper to send and retrieve messages for specific objects
type MessageHandler struct {
	Client   client.Client
	Database string
}

// Message - describes a message
type Message struct {
	Type    uint   `json:"type"`
	Message string `json:"message"`
}

// CreateMessageHandler - creates a new messagehandler instance
func CreateMessageHandler(client client.Client, database string) MessageHandler {
	return MessageHandler{
		Client:   client,
		Database: database,
	}
}

// Store - writes a single message with the given id to the store
func (m *MessageHandler) Store(id string, msg Message) {
	bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  m.Database,
		Precision: "ms",
	})

	// Create a point and add to batch
	tags := map[string]string{"log": id}
	fields := map[string]interface{}{
		"type": msg.Type,
		"msg":  msg.Message,
	}
	pt, err := client.NewPoint(id, tags, fields, time.Now())
	if err != nil {
		fmt.Println("Error: ", err.Error())
	}
	bp.AddPoint(pt)

	err = m.Client.Write(bp)
	if err != nil {
		fmt.Println("Error: ", err.Error())
	}
}

func getColumnIndex(columns []string, name string) int {
	for idx, column := range columns {
		if column == name {
			return idx
		}
	}
	return -1
}

// GetAll - retrieves all messages for the given id from the store
func (m *MessageHandler) GetAll(id string) []Message {
	q := client.NewQuery("SELECT * FROM \""+m.Database+"\".\"autogen\".\""+id+"\"", id, "ms")
	if response, err := m.Client.Query(q); err == nil && response.Error() == nil {
		if response.Error() != nil {
			fmt.Println("Error: ", response.Error().Error())
			return []Message{}
		}

		var res []Message
		for _, result := range response.Results {
			for _, series := range result.Series {
				// we enforce a panic if typeIdx or msgIdx wasnt found
				typeIdx := getColumnIndex(series.Columns, "type")
				msgIdx := getColumnIndex(series.Columns, "msg")
				for _, value := range series.Values {
					typeVal, _ := value[typeIdx].(json.Number).Int64()
					res = append(res, Message{
						Type:    uint(typeVal),
						Message: value[msgIdx].(string),
					})
				}
			}
		}
		return res
	}

	return []Message{}
}
