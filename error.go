package d2protocolbuilder

import "fmt"

type protocolError struct {
	err error
	msg string
}

func newError(err error, msg string) error {
	return &protocolError{err, msg}
}

func (e *protocolError) Error() string {
	return fmt.Sprintf("d2protocolbuilder error: %v (%v)", e.msg, e.err)
}
