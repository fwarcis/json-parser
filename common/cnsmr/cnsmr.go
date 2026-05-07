package cnsmr

import (
	"errors"
	"fmt"
	"io"

	"json-parser/common/errs"
)

var ErrCondition = errors.New("condition error")

type ConsumerError struct {
	FunctionName string
	Char         rune
	Position     int
	Iteration    int

	message string
	wrapped error
}

func NewConsumerError(
	wrapped error,
	pos int,
	iters int,
	char rune,
	funcName string,
) *ConsumerError {
	return &ConsumerError{
		funcName,
		char,
		pos,
		iters,
		"condition error: " +
			funcName +
			fmt.Sprintf(
				": '%c' at %d on %d",
				char, pos, iters),
		wrapped,
	}
}

func (e *ConsumerError) Error() string {
	return e.message
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
	pos  int
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
	return &RuneConsumer{inp, outp, 0, char, nil}, nil
}

func (c *RuneConsumer) EqualOnce(str string) error {
	if str == "" {
		c.err = errors.Join(c.err,
			errs.NewZeroValue(str, "EqualOnce", "can"))
	}
	if c.err == io.EOF {
		c.err = NewConsumerError(io.ErrUnexpectedEOF,
			c.pos, 0, c.char, "EqualOnce")
		return c.err
	} else if c.err != nil {
		return fmt.Errorf("inherited error: %w", c.err)
	}

	i := 0
	runes := []rune(str)
	length := len(runes)
	for ch := runes[0]; i < length; ch = runes[i] {
		if ch != c.char {
			c.pos += i
			c.err = NewConsumerError(ErrCondition,
				c.pos, i, c.char, "AtMin")
			return c.err
		}
		_, c.err = c.w.WriteRune(c.char)
		if c.err != nil {
			c.pos += i
			c.err = NewConsumerError(c.err,
				c.pos, i, c.char, "AtMin")
			return c.err
		}
		i++
		c.char, _, c.err = c.r.ReadRune()
		if c.err == io.EOF {
			c.pos += i
			c.err = NewConsumerError(io.ErrUnexpectedEOF,
				c.pos, i, c.char, "AtMin")
			return c.err
		} else if c.err != nil {
			c.pos += i
			c.err = NewConsumerError(c.err,
				c.pos, i, c.char, "AtMin")
			return c.err
		}
	}

	return nil
}

func (c *RuneConsumer) AtMin(count int, can func(rune) bool) error {
	if count <= 0 {
		c.err = errors.Join(c.err,
			errs.NewInvalidNumber(count, "AtMin", "count <= 0"))
	}
	if can == nil {
		c.err = errors.Join(c.err,
			errs.NewZeroValue(can, "AtMin", "can"))
	}
	if c.err == io.EOF {
		c.err = NewConsumerError(io.ErrUnexpectedEOF,
			c.pos, 0, c.char, "AtMin")
		return c.err
	} else if c.err != nil {
		return fmt.Errorf("inherited error: %w", c.err)
	}

	i := 0
	for {
		if !can(c.char) {
			c.pos += i
			c.err = NewConsumerError(ErrCondition,
				c.pos, i, c.char, "AtMin")
			return c.err
		}
		_, c.err = c.w.WriteRune(c.char)
		if c.err != nil {
			c.pos += i
			c.err = NewConsumerError(c.err,
				c.pos, i, c.char, "AtMin")
			return c.err
		}
		i++
		c.char, _, c.err = c.r.ReadRune()
		if i < count {
			if c.err == io.EOF {
				c.pos += i
				c.err = NewConsumerError(io.ErrUnexpectedEOF,
					c.pos, i, c.char, "AtMin")
				return c.err
			} else if c.err != nil {
				c.pos += i
				c.err = NewConsumerError(c.err,
					c.pos, i, c.char, "AtMin")
				return c.err
			}
		} else {
			if c.err == io.EOF {
				c.pos += i
				return io.EOF
			} else if c.err != nil {
				c.pos += i
				c.err = NewConsumerError(c.err,
					c.pos, i, c.char, "AtMin")
				return c.err
			}
			c.pos += i
			break
		}
	}

	for can(c.char) {
		_, c.err = c.w.WriteRune(c.char)
		if c.err != nil {
			c.pos += i
			c.err = NewConsumerError(c.err,
				c.pos, i, c.char, "AtMin")
			return c.err
		}
		i++
		c.char, _, c.err = c.r.ReadRune()
		if c.err == io.EOF {
			c.pos += i
			return io.EOF
		} else if c.err != nil {
			c.pos += i
			c.err = NewConsumerError(c.err,
				c.pos, i, c.char, "AtMin")
			return c.err
		}
	}

	c.pos += i
	return nil
}

func (c *RuneConsumer) Exact(count int, can func(rune) bool) error {
	if count <= 0 {
		c.err = errors.Join(c.err,
			errs.NewInvalidNumber(count, "Exact", "count <= 0"))
	}
	if can == nil {
		c.err = errors.Join(c.err,
			errs.NewZeroValue(can, "Exact", "can"))
	}
	if c.err == io.EOF {
		c.err = NewConsumerError(io.ErrUnexpectedEOF,
			c.pos, 0, c.char, "Exact")
		return c.err
	} else if c.err != nil {
		return fmt.Errorf("inherited error: %w", c.err)
	}

	for i := 0; ; {
		ok := can(c.char)
		if i == count {
			if ok {
				c.pos += i
				c.err = NewConsumerError(ErrCondition,
					c.pos, i, c.char, "Exact")
				return c.err
			}
			c.pos += i
			return nil
		}
		if !ok {
			c.pos += i
			c.err = NewConsumerError(ErrCondition,
				c.pos, i, c.char, "Exact")
			return c.err
		}
		_, c.err = c.w.WriteRune(c.char)
		if c.err != nil {
			c.pos += i
			c.err = NewConsumerError(c.err,
				c.pos, i, c.char, "Exact")
			return c.err
		}
		i++
		c.char, _, c.err = c.r.ReadRune()
		if c.err == io.EOF {
			if i == count {
				c.pos += i
				return io.EOF
			}
			c.pos += i
			c.err = NewConsumerError(io.ErrUnexpectedEOF,
				c.pos, i, c.char, "Exact")
			return c.err
		} else if c.err != nil {
			c.pos += i
			c.err = NewConsumerError(c.err,
				c.pos, i, c.char, "Exact")
			return c.err
		}
	}
}

func (c *RuneConsumer) AtMax(count int, can func(rune) bool) error {
	if count <= 0 {
		c.err = errors.Join(c.err,
			errs.NewInvalidNumber(count, "AtMax", "count <= 0"))
	}
	if can == nil {
		c.err = errors.Join(c.err,
			errs.NewZeroValue(can, "AtMax", "can"))
	}
	if c.err != nil {
		return fmt.Errorf("inherited error: %w", c.err)
	}

	for i := 0; ; {
		if can(c.char) && i == count {
			c.pos += i
			c.err = NewConsumerError(c.err,
				c.pos, i, c.char, "AtMax")
			return c.err
		} else if !can(c.char) {
			c.pos += i
			return nil
		}
		_, c.err = c.w.WriteRune(c.char)
		if c.err != nil {
			c.pos += i
			return c.err
		}
		i++
		c.char, _, c.err = c.r.ReadRune()
		if c.err == io.EOF {
			c.pos += i
			return io.EOF
		} else if c.err != nil {
			c.pos += i
			c.err = NewConsumerError(c.err,
				c.pos, i, c.char, "AtMax")
			return c.err
		}
	}
}

func (c *RuneConsumer) SkipWhile(can func(rune) bool) error {
	i := 0
	for can(c.char) {
		c.char, _, c.err = c.r.ReadRune()
		i++
		if c.err == io.EOF {
			c.pos += i
			return io.EOF
		} else if c.err != nil {
			c.pos += i
			c.err = NewConsumerError(c.err,
				c.pos, i, c.char, "SkipWhile")
			return c.err
		}
	}
	c.pos += i
	return nil
}

func (c *RuneConsumer) Reset() {
	c.err = nil
}

func (c *RuneConsumer) Char() rune {
	return c.char
}
