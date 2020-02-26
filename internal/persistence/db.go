package persistence

import (
	"encoding/json"
	"io/ioutil"

	"github.com/zebresel-com/mongodm"
)

// Connect - opens a connection to mongodb and returns the connection object
func Connect(lang string, hosts []string, dbname string, username string, password string) (*mongodm.Connection, error) {
	// try parsing a locals file
	var locals map[string]string
	if lang != "" {
		// TODO: find a decent opiniated path
		if file, err := ioutil.ReadFile("locals.json"); err == nil {
			var locale map[string]map[string]string
			err = json.Unmarshal(file, &locale)
			if err != nil {
				return nil, err
			}

			locals = locale[lang]
		}
	}

	// connect to mongodb instance
	dbConfig := &mongodm.Config{
		DatabaseHosts:    hosts,
		DatabaseName:     dbname,
		DatabaseUser:     username,
		DatabasePassword: password,
		Locals:           locals,
	}
	return mongodm.Connect(dbConfig)
}
