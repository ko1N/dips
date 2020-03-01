package database

import (
	"encoding/json"
	"io/ioutil"

	"github.com/zebresel-com/mongodm"
)

// Config - config entry describing a database config
type Config struct {
	LanguageFile string   `json:"language_file" toml:"language_file"`
	Language     string   `json:"language" toml:"language"`
	Hosts        []string `json:"hosts" toml:"hosts"`
	Database     string   `json:"database" toml:"database"`
	Username     string   `json:"username" toml:"username"`
	Password     string   `json:"password" toml:"password"`
}

// Connect - opens a connection to mongodb and returns the connection object
func Connect(conf Config) (*mongodm.Connection, error) {
	// try parsing a locals file
	var locals map[string]string
	if conf.LanguageFile != "" && conf.Language != "" {
		// TODO: find a decent opiniated path
		if file, err := ioutil.ReadFile(conf.LanguageFile); err == nil {
			var locale map[string]map[string]string
			err = json.Unmarshal(file, &locale)
			if err != nil {
				return nil, err
			}

			locals = locale[conf.Language]
		}
	}

	// connect to mongodb instance
	dbConfig := &mongodm.Config{
		DatabaseHosts:    conf.Hosts,
		DatabaseName:     conf.Database,
		DatabaseUser:     conf.Username,
		DatabasePassword: conf.Password,
		Locals:           locals,
	}
	return mongodm.Connect(dbConfig)
}
