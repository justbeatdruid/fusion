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
	alreadyBound
	permissionDenied
	unpublished
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

func AlreadyBoundError(format string, a ...interface{}) error {
	return &Error{
		t:       alreadyBound,
		message: fmt.Sprintf(format, a...),
	}
}

func PermissionDeniedError(format string, a ...interface{}) error {
	return &Error{
		t:       permissionDenied,
		message: fmt.Sprintf(format, a...),
	}
}

func UnpublishedError(format string, a ...interface{}) error {
	return &Error{
		t:       unpublished,
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

func IsAlreadyBound(err error) bool {
	if e, ok := err.(*Error); ok {
		return e.t == alreadyBound
	}
	return strings.Contains(err.Error(), "already bound")
}

func IsPermissionDenied(err error) bool {
	if e, ok := err.(*Error); ok {
		return e.t == permissionDenied
	}
	return strings.Contains(err.Error(), "permission denied")
}

func IsUnpublished(err error) bool {
	if e, ok := err.(*Error); ok {
		return e.t == unpublished
	}
	return strings.Contains(err.Error(), "unpublished")
}
