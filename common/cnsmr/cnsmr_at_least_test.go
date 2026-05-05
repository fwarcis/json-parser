package cnsmr

import (
	"bufio"
	"errors"
	"io"
	"strings"
	"testing"
)

// mockReaderWriter позволяет эмулировать ошибки на N-ной операции чтения или записи
type mockReaderWriter struct {
	readData    *strings.Reader
	readErr     error
	writeErr    error
	writesCount int
	failWriteAt int // Если -1, запись никогда не падает
}

func (m *mockReaderWriter) Read(p []byte) (int, error) {
	n, err := m.readData.Read(p)
	if err == io.EOF && m.readErr != nil && m.readErr != io.EOF {
		return n, m.readErr
	}
	return n, err
}

func (m *mockReaderWriter) Write(p []byte) (int, error) {
	if m.failWriteAt != -1 && m.writesCount >= m.failWriteAt {
		return 0, m.writeErr
	}
	m.writesCount++
	return len(p), nil
}

func TestRuneConsumer_AtLeast(t *testing.T) {
	dummyErr := errors.New("dummy generic error")
	writeErr := errors.New("write failed")

	// Вспомогательная функция для условия
	isLetter := func(r rune) bool { return r >= 'a' && r <= 'z' }
	isAlwaysTrue := func(r rune) bool { return true }

	tests := []struct {
		name        string
		n           int
		canFn       func(rune) bool
		initialChar rune
		initialErr  error

		readString  string
		readErr     error
		writeErr    error
		failWriteAt int

		wantErrText string // Какую подстроку мы ожидаем в итоговой ошибке
		wantEOF     bool   // Ожидаем ли строго io.EOF как возвращаемое значение
		wantNil     bool   // Ожидаем ли строго nil
	}{
		// --- БЛОК 1: Валидация входных параметров ---
		{
			name:        "n <= 0 without previous error",
			n:           0,
			initialChar: 'a',
			wantErrText: "n <= 0",
		},
		{
			name:        "n <= 0 with previous error",
			n:           -1,
			initialErr:  dummyErr,
			wantErrText: "dummy generic error", // Join должен сохранить обе ошибки
		},
		{
			name:        "n > 0 but existing error",
			n:           1,
			initialErr:  dummyErr,
			wantErrText: "dummy generic error",
		},

		// --- БЛОК 2: Первый цикл (Обязательные символы) ---
		{
			name:        "First loop: condition !can immediately",
			n:           2,
			canFn:       isLetter,
			initialChar: '1', // не буква
			wantErrText: "!can",
		},
		{
			name:        "First loop: write error",
			n:           2,
			canFn:       isLetter,
			initialChar: 'a',
			failWriteAt: 0,
			writeErr:    writeErr,
			wantErrText: "write failed",
		},
		{
			name:        "First loop: unexpected EOF during ReadRune",
			n:           3,
			canFn:       isLetter,
			initialChar: 'a',
			readString:  "b", // Хватит только на 1 чтение (итого 2 символа), а нужно 3
			failWriteAt: -1,
			wantErrText: "unexpected end",
		},
		{
			name:        "First loop: standard read error",
			n:           2,
			canFn:       isLetter,
			initialChar: 'a',
			readString:  "",
			readErr:     dummyErr,
			failWriteAt: -1,
			wantErrText: "dummy generic error",
		},
		{
			name:        "First loop: exact n reached with EOF",
			n:           2,
			canFn:       isLetter,
			initialChar: 'a',
			readString:  "b",
			failWriteAt: -1,
			wantEOF:     true, // Должен вернуть строго io.EOF, прервав внешний цикл
		},
		{
			name:        "First loop: exact n reached with generic read error",
			n:           2,
			canFn:       isLetter,
			initialChar: 'a',
			readString:  "b",
			readErr:     dummyErr, // При чтении после достижения n произошла ошибка
			failWriteAt: -1,
			wantErrText: "dummy generic error",
		},

		// --- БЛОК 3: Второй цикл (Опциональные символы) ---
		{
			name:        "Second loop: stop consuming gracefully (nil)",
			n:           1,
			canFn:       isLetter,
			initialChar: 'a',
			readString:  "123", // '1' остановит цикл
			failWriteAt: -1,
			wantNil:     true,
		},
		{
			name:        "Second loop: write error",
			n:           1,
			canFn:       isAlwaysTrue,
			initialChar: 'a',
			readString:  "b",
			failWriteAt: 1, // Упадет на записи второго символа ('b')
			writeErr:    writeErr,
			wantErrText: "write failed",
		},
		{
			name:        "Second loop: success via EOF",
			n:           1,
			canFn:       isAlwaysTrue,
			initialChar: 'a',
			readString:  "bc", // Прочитает 'b', потом 'c', потом упрется в EOF
			failWriteAt: -1,
			wantEOF:     true,
		},
		{
			name:        "Second loop: standard read error",
			n:           1,
			canFn:       isAlwaysTrue,
			initialChar: 'a',
			readString:  "b",
			readErr:     dummyErr,
			failWriteAt: -1,
			wantErrText: "dummy generic error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Настройка моков потоков ввода-вывода
			mockIO := &mockReaderWriter{
				readData:    strings.NewReader(tt.readString),
				readErr:     tt.readErr,
				writeErr:    tt.writeErr,
				failWriteAt: tt.failWriteAt,
			}

			// Буфер размером 1 байт для отключения кэширования bufio
			// (иначе bufio.Reader скроет ошибку до опустошения буфера)
			reader := bufio.NewReaderSize(mockIO, 1)
			writer := bufio.NewWriterSize(mockIO, 1)

			c := &RuneConsumer{
				r:    reader,
				w:    writer,
				char: tt.initialChar,
				err:  tt.initialErr,
			}

			can := tt.canFn
			if can == nil {
				can = func(r rune) bool { return true }
			}

			err := c.AtLeast(tt.n, can)

			// Если мы писали через bufio.Writer, нужно сбросить буфер,
			// чтобы триггернуть замоканную ошибку записи (если она не стрельнула на WriteRune)
			if err == nil {
				err = writer.Flush()
				c.err = err // симулируем поведение консьюмера
			}

			// Проверки ожиданий
			if tt.wantNil {
				if err != nil {
					t.Fatalf("expected nil error, got: %v", err)
				}
				return
			}

			if tt.wantEOF {
				if !errors.Is(err, io.EOF) {
					t.Fatalf("expected strictly io.EOF, got: %v", err)
				}
				return
			}

			if err == nil {
				t.Fatalf("expected an error containing %q, but got nil", tt.wantErrText)
			}

			if !strings.Contains(err.Error(), tt.wantErrText) {
				t.Errorf("error %q does not contain expected substring %q", err.Error(), tt.wantErrText)
			}
		})
	}
}
