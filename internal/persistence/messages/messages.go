package messages

import (
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

// CreateMessageHandler - creates a new messagehandler instance
func CreateMessageHandler(client client.Client, database string) MessageHandler {
	return MessageHandler{
		Client:   client,
		Database: database,
	}
}

// Store - writes a single message with the given id to the store
func (m *MessageHandler) Store(id string, msg string) {
	bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  m.Database,
		Precision: "ms",
	})

	// Create a point and add to batch
	tags := map[string]string{"log": id}
	fields := map[string]interface{}{
		"msg": msg,
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

// GetAll - retrieves all messages for the given id from the store
func (m *MessageHandler) GetAll(id string) []string {
	q := client.NewQuery("SELECT * FROM \""+m.Database+"\".\"autogen\".\""+id+"\"", id, "ms")
	if response, err := m.Client.Query(q); err == nil && response.Error() == nil {
		if response.Error() != nil {
			fmt.Println("Error: ", response.Error().Error())
			return []string{}
		}

		var res []string
		for _, result := range response.Results {
			for _, series := range result.Series {
				for _, value := range series.Values {
					res = append(res, value[2].(string))
				}
			}
		}
		return res
	}
	return []string{}
}
