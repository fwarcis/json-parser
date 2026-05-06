package cnsmr

import (
	"errors"
	"fmt"
	"io"

	"json-parser/common/errs"
)

var ErrBeforeCondition = &ConsumerError{nil, ""}

func NewBeforeScanError(format string, args ...any) *ConsumerError {
	return NewConsumerError(ErrBeforeCondition, format, args...)
}

var ErrCondition = &ConsumerError{nil, ""}

func NewConditionError(format string, args ...any) *ConsumerError {
	return NewConsumerError(ErrCondition, format, args...)
}

var ErrOnWrite = &ConsumerError{nil, ""}

func NewOnWriteError(format string, args ...any) *ConsumerError {
	return NewConsumerError(ErrOnWrite, format, args...)
}

var ErrOnRead = &ConsumerError{nil, ""}

func NewOnReadError(format string, args ...any) *ConsumerError {
	return NewConsumerError(ErrOnRead, format, args...)
}

func IsSuccess(err error) bool {
	return err == nil || err == io.EOF
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

type RuneWriter interface {
	WriteRune(r rune) (size int, err error)
}

type RuneConsumer struct {
	r    io.RuneReader
	w    RuneWriter
	char rune
	err  error
}

func NewRuneConsumer(
	inp io.RuneReader,
	outp RuneWriter,
) (*RuneConsumer, error) {
	char, _, err := inp.ReadRune()
	if err != nil {
		return nil, err
	}
	return &RuneConsumer{inp, outp, char, nil}, nil
}

func (c *RuneConsumer) AtMin(n int, can func(rune) bool) error {
	if n <= 0 {
		if c.err == io.EOF {
			c.err = errors.Join(io.ErrUnexpectedEOF,
				NewBeforeScanError(
					"AtMin: '%c': unexpected end (n=%d)",
					c.char, n),
				errs.NewInvalidNumber(
					"AtMin: '%c': n <= 0 (n=%d)", c.char, n))
		} else if c.err != nil {
			c.err = errors.Join(c.err,
				NewBeforeScanError(
					"AtMin: '%c': (n=%d)",
					c.char, n),
				errs.NewInvalidNumber(
					"AtMin: '%c': n <= 0 (n=%d)", c.char, n))
		} else {
			c.err = errors.Join(
				NewBeforeScanError(
					"AtMin: '%c': (n=%d)",
					c.char, n),
				errs.NewInvalidNumber(
					"AtMin: '%c': n <= 0 (n=%d)", c.char, n))
		}
		return c.err
	}
	if c.err == io.EOF {
		c.err = errors.Join(io.ErrUnexpectedEOF,
			NewBeforeScanError(
				"AtMin: '%c': unexpected end (n=%d)",
				c.char, n))
		return c.err
	} else if c.err != nil {
		c.err = errors.Join(c.err,
			NewBeforeScanError(
				"AtMin: '%c': (n=%d)",
				c.char, n))
		return c.err
	}

	for i := 0; ; {
		if !can(c.char) {
			c.err = NewConditionError(
				"AtMin: '%c': condition failed (n=%d, i=%d)",
				c.char, n, i)
			return c.err
		}
		_, c.err = c.w.WriteRune(c.char)
		if c.err != nil {
			c.err = errors.Join(c.err,
				NewOnWriteError(
					"AtMin: '%c' (n=%d, i=%d)",
					c.char, n, i))
			return c.err
		}
		i++
		c.char, _, c.err = c.r.ReadRune()
		if i < n {
			if c.err == io.EOF {
				c.err = errors.Join(io.ErrUnexpectedEOF,
					NewOnReadError(
						"AtMin: '%c': unexpected end (n=%d, i=%d)",
						c.char, n, i))
				return c.err
			} else if c.err != nil {
				c.err = errors.Join(c.err,
					NewOnReadError(
						"AtMin: '%c' (n=%d, i=%d)",
						c.char, n, i))
				return c.err
			}
		} else {
			if c.err == io.EOF {
				return io.EOF
			} else if c.err != nil {
				c.err = errors.Join(c.err,
					NewOnReadError(
						"AtMin: '%c' (n=%d, i=%d)",
						c.char, n, i))
				return c.err
			}
			break
		}
	}

	for i := 0; can(c.char); {
		_, c.err = c.w.WriteRune(c.char)
		if c.err != nil {
			c.err = errors.Join(c.err, NewOnWriteError(
				"AtMin: '%c' (i=%d)", c.char, i))
			return c.err
		}
		i++
		c.char, _, c.err = c.r.ReadRune()
		if c.err == io.EOF {
			return io.EOF
		} else if c.err != nil {
			c.err = errors.Join(c.err, NewOnReadError(
				"AtMin: '%c' (i=%d)", c.char, i))
			return c.err
		}
	}

	return nil
}

func (c *RuneConsumer) Exact(n int, can func(rune) bool) error {
	if n <= 0 {
		if c.err == io.EOF {
			c.err = errors.Join(io.ErrUnexpectedEOF,
				NewBeforeScanError(
					"Only: '%c': unexpected end (n=%d)",
					c.char, n),
				errs.NewInvalidNumber(
					"Only: '%c': n <= 0 (n=%d)", c.char, n))
		} else if c.err != nil {
			c.err = errors.Join(c.err,
				NewBeforeScanError(
					"Only: '%c': (n=%d)",
					c.char, n),
				errs.NewInvalidNumber(
					"Only: '%c': n <= 0 (n=%d)", c.char, n))
		} else {
			c.err = errors.Join(
				NewBeforeScanError(
					"Only: '%c': (n=%d)",
					c.char, n),
				errs.NewInvalidNumber(
					"Only: '%c': n <= 0 (n=%d)", c.char, n))
		}
		return c.err
	}
	if c.err == io.EOF {
		c.err = errors.Join(io.ErrUnexpectedEOF,
			NewBeforeScanError(
				"Only: '%c': unexpected end (n=%d)",
				c.char, n))
		return c.err
	} else if c.err != nil {
		c.err = errors.Join(c.err,
			NewBeforeScanError(
				"Only: '%c': (n=%d)",
				c.char, n))
		return c.err
	}

	for i := 0; ; {
		ok := can(c.char)
		if i == n {
			if ok {
				c.err = NewConditionError(
					"Only: '%c': condition failed (n=%d, i=%d)",
					c.char, n, i)
				return c.err
			}
			return nil
		}
		if !ok {
			c.err = NewConditionError(
				"Only: '%c': condition failed (n=%d, i=%d)",
				c.char, n, i)
			return c.err
		}
		_, c.err = c.w.WriteRune(c.char)
		if c.err != nil {
			c.err = errors.Join(c.err,
				NewOnWriteError(
					"Only: '%c' (n=%d, i=%d)",
					c.char, n, i))
			return c.err
		}
		i++
		c.char, _, c.err = c.r.ReadRune()
		if c.err == io.EOF {
			if i == n {
				return io.EOF
			}
			c.err = errors.Join(io.ErrUnexpectedEOF,
				NewOnReadError(
					"Only: '%c': unexpected end (n=%d, i=%d)",
					c.char, n, i))
			return c.err
		} else if c.err != nil {
			c.err = errors.Join(c.err,
				NewOnReadError(
					"Only: '%c' (n=%d, i=%d)",
					c.char, n, i))
			return c.err
		}
	}
}

func (c *RuneConsumer) AtMax(n int, can func(rune) bool) error {
	if n <= 0 {
		if c.err == io.EOF {
			c.err = errors.Join(
				NewBeforeScanError(
					"AtMax: '%c': (n=%d)",
					c.char, n),
				errs.NewInvalidNumber(
					"AtMax: '%c': n <= 0 (n=%d)", c.char, n))
		} else {
			c.err = errors.Join(c.err,
				NewBeforeScanError(
					"AtMax: '%c': (n=%d)",
					c.char, n),
				errs.NewInvalidNumber(
					"AtMax: '%c': n <= 0 (n=%d)", c.char, n))
		}
		return c.err
	}
	if c.err == io.EOF {
		return io.EOF
	} else if c.err != nil {
		c.err = errors.Join(c.err,
			NewBeforeScanError(
				"AtMax: '%c': (n=%d)",
				c.char, n))
		return c.err
	}

	for i := 0; ; {
		if can(c.char) && i == n {
			c.err = NewConditionError(
				"AtMax: '%c': condition failed (n=%d, i=%d)",
				c.char, n, i)
			return c.err
		} else if !can(c.char) {
			return nil
		}
		_, c.err = c.w.WriteRune(c.char)
		if c.err != nil {
			c.err = errors.Join(c.err,
				NewOnWriteError(
					"AtMax: '%c' (n=%d, i=%d)",
					c.char, n, i))
			return c.err
		}
		i++
		c.char, _, c.err = c.r.ReadRune()
		if c.err == io.EOF {
			return io.EOF
		} else if c.err != nil {
			c.err = errors.Join(c.err,
				NewOnReadError(
					"AtMax: '%c' (n=%d, i=%d)",
					c.char, n, i))
			return c.err
		}
	}
}

func (c *RuneConsumer) Char() rune {
	return c.char
}

func (c *RuneConsumer) Err() error {
	return c.err
}

func (c *RuneConsumer) Reset() {
	c.err = nil
}
