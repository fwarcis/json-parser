package cnsmr

import (
	"errors"
	"io"
	"strings"
	"testing"
)

// --- Tests for Logic Flaws ---

func TestRuneConsumer_RecoveryAfterConditionFailure(t *testing.T) {
	// CONTRACT: A ConditionError (predicate mismatch) should notify the caller
	// that the rule didn't match, but it should NOT kill the consumer.
	// The caller should be able to try a different rule on the same character.

	input := "b"
	writer := &MockWriter{}
	c, _ := NewRuneConsumer(strings.NewReader(input), writer)

	// Attempt 1: Try to consume 'a'. This should fail the condition.
	err1 := c.AtLeast(1, func(r rune) bool { return r == 'a' })
	if !errors.Is(err1, ErrCondition) {
		t.Errorf("Expected ErrCondition, got %v", err1)
	}

	// Attempt 2: Try to consume 'b'. Since 'b' hasn't been consumed yet,
	// this should succeed.
	// BUG: This will FAIL because c.err is "sticky" and returns the previous error.
	err2 := c.AtLeast(1, func(r rune) bool { return r == 'b' })
	if !errors.Is(err2, io.EOF) {
		t.Errorf("Expected success (nil or io.EOF) on second attempt, but consumer is poisoned: %v", err2)
	}

	if writer.buf.String() != "b" {
		t.Errorf("Expected 'b' to be written after recovery, got %q", writer.buf.String())
	}
}

func TestRuneConsumer_AtLeastZero(t *testing.T) {
	// CONTRACT: "At least 0" is mathematically always true (a no-op).
	// It should succeed immediately without changing state or writing.

	c, _ := NewRuneConsumer(strings.NewReader("abc"), &MockWriter{})

	// BUG: The current implementation explicitly returns an error for n <= 0.
	err := c.AtLeast(0, func(r rune) bool { return false })
	if err != nil {
		t.Errorf("Expected AtLeast(0) to be a successful no-op, got error: %v", err)
	}
}

func TestRuneConsumer_EOFIsSuccessNotPoison(t *testing.T) {
	// CONTRACT: io.EOF is a signal that the stream ended successfully.
	// Subsequent calls that require 0 characters should still succeed.

	input := "a"
	writer := &MockWriter{}
	c, _ := NewRuneConsumer(strings.NewReader(input), writer)

	// Consume the only character
	err := c.AtLeast(1, func(r rune) bool { return r == 'a' })
	if !errors.Is(err, io.EOF) {
		t.Fatalf("Expected io.EOF on completion, got %v", err)
	}

	// BUG: Because io.EOF was stored in c.err, this valid call (n=0 or similar)
	// will now return an error joined with io.EOF instead of being a success.
	// Also, in the current code, n=0 is an error anyway, but even if fixed,
	// the sticky EOF would prevent further logic.
	err2 := c.AtLeast(0, func(r rune) bool { return true })
	if errors.Is(err2, errInvalidNumber) { // Check if it's hitting the validation logic
		t.Errorf("AtLeast(0) should not be blocked by a previous successful EOF")
	}
}

func TestRuneConsumer_InconsistentReturnValues(t *testing.T) {
	// CONTRACT: If the mandatory condition (n) is met, the function should
	// ideally return the same success signal regardless of whether it stopped
	// because of a non-matching character or the end of the stream.

	alwaysA := func(r rune) bool { return r == 'a' }

	t.Run("StoppedByMismatch", func(t *testing.T) {
		c, _ := NewRuneConsumer(strings.NewReader("ab"), &MockWriter{})
		err := c.AtLeast(1, alwaysA) // Should consume 'a', stop at 'b'
		// Current code returns: nil
		if err != nil {
			t.Errorf("Expected nil on mismatch stop, got %v", err)
		}
	})

	t.Run("StoppedByEOF", func(t *testing.T) {
		c, _ := NewRuneConsumer(strings.NewReader("a"), &MockWriter{})
		err := c.AtLeast(1, alwaysA) // Should consume 'a', stop at EOF
		// BUG: Current code returns: io.EOF
		// This inconsistency forces the caller to check: if err != nil && err != io.EOF
		if err != nil {
			t.Errorf("Inconsistency: Expected nil for all successes, but EOF return is %v", err)
		}
	})
}

func TestRuneConsumer_WriterFailureDoesNotConsume(t *testing.T) {
	// CONTRACT: If the writer fails, the character was not successfully "consumed".
	// It should remain as the current 'c.char' so it can be retried if the error is transient.

	writer := &MockWriter{err: errors.New("temporary failure")}
	c, _ := NewRuneConsumer(strings.NewReader("a"), writer)

	err := c.AtLeast(1, func(r rune) bool { return r == 'a' })
	if err == nil {
		t.Fatal("Expected write error")
	}

	// Fix the writer
	writer.err = nil

	// BUG: Even if we ignore the sticky error bug, the code's loop structure
	// does not clearly define if c.char is reset or advanced on write failures.
	// In the current code, it returns before ReadRune, so c.char is actually still 'a'.
	// However, the sticky error c.err = Join(...) prevents this retry from ever working.
	err = c.AtLeast(1, func(r rune) bool { return r == 'a' })
	if err != nil {
		t.Errorf("Retry after fixing writer should succeed, but got: %v", err)
	}
}
