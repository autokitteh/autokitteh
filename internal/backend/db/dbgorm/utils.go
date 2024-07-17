package dbgorm

import (
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func schemaToSDK[M any, O sdktypes.Object](model *M, err error, parseFunc func(model M) (O, error)) (O, error) {
	var invalid O
	if model == nil || err != nil {
		return invalid, translateError(err)
	}
	return parseFunc(*model)
}

func schemasToSDK[M any, O sdktypes.Object](models []M, err error, parseFunc func(model M) (O, error)) ([]O, error) {
	if models == nil || err != nil {
		return nil, translateError(err)
	}
	return kittehs.TransformError(models, parseFunc)
}
