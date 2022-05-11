package jsonapi

import (
	"context"
	"io"

	"github.com/ebladrocher/keypass/storepass"
)

// API ...
type API struct {
	Store  *storepass.RootStore
	Reader io.Reader
	Writer io.Writer
}

// ReadAndRespond ...
func (api *API) ReadAndRespond(ctx context.Context) error {
	message, err := readMessage(api.Reader)
	if message == nil || err != nil {
		return err
	}

	return api.respondMessage(ctx, message)
}

// RespondError ...
func (api *API) RespondError(err error) error {
	var response errorResponse
	response.Error = err.Error()

	return sendSerializedJSONMessage(response, api.Writer)
}
