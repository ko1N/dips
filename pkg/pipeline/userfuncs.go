package pipeline

import (
	"github.com/d5/tengo/v2"
	"github.com/google/uuid"
)

var userFunctions = map[string](func(args ...tengo.Object) (tengo.Object, error)){
	"uuid": tuuid,
}

func tuuid(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 0 {
		return nil, tengo.ErrWrongNumArguments
	}

	u := uuid.New().String()
	return tengo.FromInterface(u)
}
