package error

import "errors"

var (
	ErrReadingBody = errors.New("error reading body")
)

type (
	HandlerError struct {
		Cause error
	}
)

func (e HandlerError) Error() string {
	return e.Cause.Error()
}
