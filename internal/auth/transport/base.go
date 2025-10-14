package transport

import (
	"fmt"
	"net/http"
)

type NoTokenError struct {
	Err error
}

func (e *NoTokenError) Error() string {
	return fmt.Sprintf("no token: %v", e.Err)

}

func (e *NoTokenError) Unwrap() error {
	return e.Err
}

type Transport interface {
	Read(*http.Request) (string, error)
	Write(http.ResponseWriter, string) error
}
