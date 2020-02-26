package persistence

import (
	"errors"
	"reflect"
	"strings"

	"github.com/zebresel-com/mongodm"
)

// CrudWrapper - Represents a crud wrapper and all required data
type CrudWrapper struct {
	db             *mongodm.Connection
	typeName       string
	collectionName string
}

func getCollectionName(typeName string) string {
	return strings.ToLower(typeName) + "s"
}

func getTypeName(val interface{}) string {
	if t := reflect.TypeOf(val); t.Kind() == reflect.Ptr {
		return t.Elem().Name()
	} else {
		return t.Name()
	}
}

// CreateCrudWrapper - Creates a new CRUD wrapper for the given types
func CreateCrudWrapper(db *mongodm.Connection, document mongodm.IDocumentBase) (CrudWrapper, error) {
	//fmt.Printf("name: %+v\n", getType(document))

	typeName := getTypeName(document)
	if typeName == "" {
		return CrudWrapper{}, errors.New("could not resolve document type name")
	}

	collectionName := getCollectionName(typeName)
	db.Register(document, collectionName)

	return CrudWrapper{
		db:             db,
		typeName:       typeName,
		collectionName: collectionName,
	}, nil
}

// Create - creates a new document
func (c *CrudWrapper) Create() {
}
