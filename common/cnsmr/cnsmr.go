package cnsmr

import (
	"bufio"
	"errors"
	"fmt"
	"io"

	"json-parser/common/errs"
)

var ErrCondition = &ConsumerError{nil, ""}

func NewConditionError(format string, args ...any) *ConsumerError {
	return NewConsumerError(ErrCondition, format, args...)
}

var ErrOnConsume = &ConsumerError{nil, ""}

func NewOnConsumeError(format string, args ...any) *ConsumerError {
	return NewConsumerError(ErrOnConsume, format, args...)
}

var ErrNextChar = &ConsumerError{nil, ""}

func NewNextCharError(format string, args ...any) *ConsumerError {
	return NewConsumerError(ErrNextChar, format, args...)
}

type ConsumerError struct {
	wrapped error
	message string
}

func NewConsumerError(
	wrapped error, format string, args ...any,
) *ConsumerError {
	return &ConsumerError{
		wrapped: wrapped,
		message: fmt.Sprintf(format, args...),
	}
}

func (e *ConsumerError) Error() string {
	return "consumer error: " + e.message
}

func (e *ConsumerError) Unwrap() error {
	return e.wrapped
}

type RuneConsumer struct {
	r    *bufio.Reader
	w    *bufio.Writer
	char rune
	err  error
}

func NewRuneConsumer(
	in io.Reader,
	out io.Writer,
) (*RuneConsumer, error) {
	r := bufio.NewReader(in)
	char, _, err := r.ReadRune()
	if err != nil {
		return nil, err
	}
	return &RuneConsumer{
		r,
		bufio.NewWriter(out),
		char,
		nil,
	}, nil
}

// Success on io.EOF | nil
func (c *RuneConsumer) AtLeast(n int, can func(rune) bool) error {
	if n <= 0 {
		if c.err != nil {
			c.err = errors.Join(c.err, errs.NewInvalidNumber(
				"AtLeast: '%c': n <= 0 (n=%d)", c.char, n))
		} else {
			c.err = errs.NewInvalidNumber(
				"AtLeast: '%c': n <= 0 (n=%d)", c.char, n)
		}
		return c.err
	}
	if c.err != nil {
		return c.err
	}

	for i := 0; ; {
		if !can(c.char) {
			c.err = NewConditionError(
				"AtLeast: '%c': !can (n=%d, i=%d)",
				c.char, n, i)
			return c.err
		}
		_, c.err = c.w.WriteRune(c.char)
		if c.err != nil {
			c.err = errors.Join(c.err,
				NewOnConsumeError(
					"AtLeast: '%c' (n=%d, i=%d)",
					c.char, n, i))
			return c.err
		}
		i++
		c.char, _, c.err = c.r.ReadRune()
		if i < n {
			if c.err == io.EOF {
				c.err = errors.Join(io.ErrUnexpectedEOF,
					NewNextCharError(
						"AtLeast: '%c': unexpected end (n=%d, i=%d)",
						c.char, n, i))
				return c.err
			} else if c.err != nil {
				c.err = errors.Join(c.err,
					NewNextCharError(
						"AtLeast: '%c' (n=%d, i=%d)",
						c.char, n, i))
				return c.err
			}
		} else {
			if c.err == io.EOF {
				return io.EOF
			} else if c.err != nil {
				c.err = errors.Join(c.err,
					NewNextCharError(
						"AtLeast: '%c' (n=%d, i=%d)",
						c.char, n, i))
				return c.err
			}
			break
		}
	}

	for i := 0; can(c.char); {
		_, c.err = c.w.WriteRune(c.char)
		if c.err != nil {
			c.err = errors.Join(c.err, NewOnConsumeError(
				"AtLeast: '%c' (i=%d)", c.char, i))
			return c.err
		}
		i++
		c.char, _, c.err = c.r.ReadRune()
		if c.err == io.EOF {
			return io.EOF
		} else if c.err != nil {
			c.err = errors.Join(c.err, NewNextCharError(
				"AtLeast: '%c' (i=%d)", c.char, i))
			return c.err
		}
	}

	return nil
}

func (c *RuneConsumer) Err() error {
	return c.err
}
