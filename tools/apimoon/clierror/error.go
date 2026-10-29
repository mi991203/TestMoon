package clierror

import (
	"errors"
	"strings"
)

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
	Status  int    `json:"http_status,omitempty"`
}

func (e *Error) Error() string        { return e.Message }
func New(code, message string) *Error { return &Error{Code: code, Message: message} }
func From(err error) *Error {
	var e *Error
	if errors.As(err, &e) {
		return e
	}
	return New("local", err.Error())
}
func Redact(s string, secrets ...string) string {
	for _, secret := range secrets {
		if secret != "" {
			s = strings.ReplaceAll(s, secret, "***")
		}
	}
	return s
}
