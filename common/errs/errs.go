package errs

import (
	"fmt"
)

type Number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64
}

func NewZeroValue[V any](
	val V, prefix, name string,
) *ValueError[V] {
	return NewValue(
		val,
		prefix,
		name+" == ZeroValue"+fmt.Sprintf("(%T)", val))
}

func NewInvalidNumber[V Number](
	val V, prefix, format string, args ...any,
) *ValueError[V] {
	return NewValue(
		val,
		prefix,
		format+fmt.Sprintf(" (%v)", val),
		args...)
}

type ValueError[V any] struct {
	Value   V
	message string
}

func NewValue[V any](
	val V, prefix, format string, args ...any,
) *ValueError[V] {
	return &ValueError[V]{
		val,
		"value error: " + prefix + ": " + fmt.Sprintf(format, args...),
	}
}

func (e *ValueError[V]) Error() string {
	return e.message
}
