package cnsmr

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
)

// --- Mocks & Helpers ---

// MockWriter allows us to inject errors into the WriteRune method.
type MockWriter struct {
	err error
	buf bytes.Buffer
}

func (m *MockWriter) WriteRune(r rune) (int, error) {
	if m.err != nil {
		return 0, m.err
	}
	return m.buf.WriteRune(r)
}

// Mock of the errs package logic mentioned in the source code
var errInvalidNumber = errors.New("invalid number")

func isInvalidNumber(err error) bool {
	// In a real scenario, this would check the specific type from the 'errs' package
	return strings.Contains(err.Error(), "n <= 0")
}

// --- Test Suite ---

func TestNewRuneConsumer(t *testing.T) {
	t.Run("InitializeWithEmptyReaderReturnsError", func(t *testing.T) {
		reader := strings.NewReader("")
		writer := &MockWriter{}
		_, err := NewRuneConsumer(reader, writer)

		if !errors.Is(err, io.EOF) {
			t.Errorf("expected io.EOF on empty reader, got %v", err)
		}
	})

	t.Run("InitializeWithContentConsumesFirstRune", func(t *testing.T) {
		reader := strings.NewReader("abc")
		writer := &MockWriter{}
		_, err := NewRuneConsumer(reader, writer)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestRuneConsumer_AtLeast_Validation(t *testing.T) {
	t.Run("ErrorWhenNIsZero", func(t *testing.T) {
		c, _ := NewRuneConsumer(strings.NewReader("abc"), &MockWriter{})
		err := c.AtMin(0, func(r rune) bool { return true })

		if err == nil || !isInvalidNumber(err) {
			t.Errorf("expected invalid number error for n=0, got %v", err)
		}
	})

	t.Run("PreservesPreviousError", func(t *testing.T) {
		// Simulating a state where c.err is already set
		c := &RuneConsumer{err: errors.New("prior error")}
		err := c.AtMin(1, func(r rune) bool { return true })

		if err == nil || !strings.Contains(err.Error(), "prior error") {
			t.Errorf("expected method to return existing consumer error, got %v", err)
		}
	})
}

func TestRuneConsumer_AtLeast_Contract(t *testing.T) {
	alwaysTrue := func(r rune) bool { return true }

	t.Run("SuccessExactlyN", func(t *testing.T) {
		input := "abc"
		writer := &MockWriter{}
		reader := strings.NewReader(input)
		c, _ := NewRuneConsumer(reader, writer)

		// Consume exactly 3
		err := c.AtMin(3, alwaysTrue)

		if !errors.Is(err, io.EOF) {
			t.Errorf("expected io.EOF after consuming all input, got %v", err)
		}
		if writer.buf.String() != "abc" {
			t.Errorf("expected 'abc' to be written, got %q", writer.buf.String())
		}
	})

	t.Run("SuccessMoreThanN", func(t *testing.T) {
		input := "aaaaa" // 5 runes
		writer := &MockWriter{}
		c, _ := NewRuneConsumer(strings.NewReader(input), writer)

		// Require at least 2, but provide 5
		err := c.AtMin(2, func(r rune) bool { return r == 'a' })

		if !errors.Is(err, io.EOF) {
			t.Errorf("expected io.EOF, got %v", err)
		}
		if writer.buf.String() != "aaaaa" {
			t.Errorf("expected all 'a's to be consumed, got %q", writer.buf.String())
		}
	})

	t.Run("FailsWhenConditionNotMetBeforeN", func(t *testing.T) {
		input := "ab"
		writer := &MockWriter{}
		c, _ := NewRuneConsumer(strings.NewReader(input), writer)

		// Require 2 'a's, but the second is 'b'
		err := c.AtMin(2, func(r rune) bool { return r == 'a' })

		if !errors.Is(err, ErrCondition) {
			t.Errorf("expected ErrCondition, got type %T (val: %v)", err, err)
		}
	})

	t.Run("StopConsumingWhenConditionFailsAfterN", func(t *testing.T) {
		input := "aaab"
		writer := &MockWriter{}
		c, _ := NewRuneConsumer(strings.NewReader(input), writer)

		// Require at least 2 'a's. Should stop at 'b' and return nil (not EOF, because 'b' is still in stream)
		err := c.AtMin(2, func(r rune) bool { return r == 'a' })
		if err != nil {
			t.Errorf("expected nil error (successful stop), got %v", err)
		}
		if writer.buf.String() != "aaa" {
			t.Errorf("expected 'aaa' to be written, got %q", writer.buf.String())
		}
	})

	t.Run("UnexpectedEOFBeforeN", func(t *testing.T) {
		input := "a"
		writer := &MockWriter{}
		c, _ := NewRuneConsumer(strings.NewReader(input), writer)

		err := c.AtMin(2, alwaysTrue)

		if !errors.Is(err, io.ErrUnexpectedEOF) {
			t.Errorf("expected io.ErrUnexpectedEOF, got %v", err)
		}
		if !errors.Is(err, ErrOnRead) {
			t.Errorf("expected error to wrap ErrNextChar, got %v", err)
		}
	})

	t.Run("HandlesWriterError", func(t *testing.T) {
		writerErr := errors.New("disk full")
		writer := &MockWriter{err: writerErr}
		c, _ := NewRuneConsumer(strings.NewReader("abc"), writer)

		err := c.AtMin(1, alwaysTrue)

		if !errors.Is(err, writerErr) {
			t.Errorf("expected original writer error, got %v", err)
		}
		if !errors.Is(err, ErrOnWrite) {
			t.Errorf("expected error to wrap ErrOnConsume, got %v", err)
		}
	})
}

func TestRuneConsumer_AtLeast_ComplexState(t *testing.T) {
	t.Run("SequentialCallsMaintainState", func(t *testing.T) {
		input := "aaabbb"
		writer := &MockWriter{}
		c, err := NewRuneConsumer(strings.NewReader(input), writer)
		if err != nil {
			t.Fatal(err)
		}

		// First call: consume 3 'a's
		err = c.AtMin(2, func(r rune) bool { return r == 'a' })
		if err != nil {
			t.Errorf("first AtLeast failed: %v", err)
		}

		// Second call: consume 3 'b's
		err = c.AtMin(2, func(r rune) bool { return r == 'b' })
		if !errors.Is(err, io.EOF) {
			t.Errorf("second AtLeast failed: expected EOF, got %v", err)
		}

		if writer.buf.String() != "aaabbb" {
			t.Errorf("expected combined output 'aaabbb', got %q", writer.buf.String())
		}
	})
}
