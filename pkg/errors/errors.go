package errors

import (
	"fmt"
	"strings"
)

type Error struct {
	t       errType
	message string
}

func (e *Error) Error() string {
	return e.message
}

type errType int

const (
	notFound errType = iota
	nameDuplicated
	contentNotVoid
)

func NotFoundError(format string, a ...interface{}) error {
	return &Error{
		t:       notFound,
		message: fmt.Sprintf(format, a...),
	}
}

func NameDuplicatedError(format string, a ...interface{}) error {
	return &Error{
		t:       nameDuplicated,
		message: fmt.Sprintf(format, a...),
	}
}

func ContentNotVoidError(format string, a ...interface{}) error {
	return &Error{
		t:       contentNotVoid,
		message: fmt.Sprintf(format, a...),
	}
}

func IsNotFound(err error) bool {
	if e, ok := err.(*Error); ok {
		return e.t == notFound
	}
	return strings.Contains(err.Error(), "not found")
}

func IsNameDuplicated(err error) bool {
	if e, ok := err.(*Error); ok {
		return e.t == nameDuplicated
	}
	return strings.Contains(err.Error(), "name duplicated")
}

func IsContentNotVoid(err error) bool {
	if e, ok := err.(*Error); ok {
		return e.t == contentNotVoid
	}
	return strings.Contains(err.Error(), "content not void")
}
