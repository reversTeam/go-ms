package core

import (
	"fmt"
	"net/http"
)

type HttpError struct {
	Code    int
	Message string
}

func NewHttpError(code int, args ...any) *HttpError {
	message := http.StatusText(code)

	if len(args) > 0 && args[0] != "" {
		if len(args) > 1 {
			message = fmt.Sprintf(args[0].(string), args[1:]...)
		} else {
			message = args[0].(string)
		}
	}

	return &HttpError{
		Code:    code,
		Message: message,
	}
}

func (e *HttpError) Error() string {
	return e.Message
}
