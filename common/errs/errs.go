package errs

import "fmt"

var ErrInvalidNumber = NewValue(nil, "invalid number")

func NewInvalidNumber(format string, args ...any) *ValueError {
	return NewValue(ErrInvalidNumber, format, args...)
}

var ErrNil = NewValue(nil, "nil value")

func NewNil(format string, args ...any) *ValueError {
	return NewValue(ErrNil, format, args...)
}

func NewValue(
	wrapped error, format string, args ...any,
) *ValueError {
	return &ValueError{
		wrapped: wrapped,
		message: fmt.Sprintf(format, args...),
	}
}

type ValueError struct {
	wrapped error
	message string
}

func (e *ValueError) Error() string {
	return "value error: " + e.message
}

func (e *ValueError) Unwrap() error {
	return e.wrapped
}
