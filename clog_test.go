package clog

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func init() {
	inTest = true
}

func TestLevel_String(t *testing.T) {
	invalidLevel := Level(-1)
	defer func() {
		assert.NotNil(t, recover())
	}()

	_ = invalidLevel.String()
}

type bufConfig struct {
	buf *bytes.Buffer
}

var _ Logger = (*bufLogger)(nil)

type bufLogger struct {
	buf *bytes.Buffer
	*noopLogger
}

func (l *bufLogger) Write(m Messager) error {
	_, err := l.buf.WriteString(m.String() + "\n")
	return err
}

func Test_memoryLogger(t *testing.T) {
	mode1 := Mode("mode1")
	level1 := LevelTrace
	NewRegister(mode1, func(v interface{}) (Logger, error) {
		cfg, ok := v.(bufConfig)
		if !ok {
			return nil, fmt.Errorf("invalid config object: want %T got %T", bufConfig{}, v)
		}
		return &bufLogger{
			buf: cfg.buf,
			noopLogger: &noopLogger{
				mode:  mode1,
				level: level1,
			},
		}, nil
	})

	mode2 := Mode("mode2")
	level2 := LevelError
	NewRegister(mode2, func(v interface{}) (Logger, error) {
		cfg, ok := v.(bufConfig)
		if !ok {
			return nil, fmt.Errorf("invalid config object: want %T got %T", &bufConfig{}, v)
		}
		return &bufLogger{
			buf: cfg.buf,
			noopLogger: &noopLogger{
				mode:  mode2,
				level: level2,
			},
		}, nil
	})

	tests := []struct {
		name         string
		fn           func(string, ...interface{})
		containsStr1 string
		containsStr2 string
	}{
		{
			name:         "trace",
			fn:           Trace,
			containsStr1: "[TRACE] log message",
			containsStr2: "",
		},
		{
			name:         "info",
			fn:           Info,
			containsStr1: "[ INFO] log message",
			containsStr2: "",
		},
		{
			name:         "warn",
			fn:           Warn,
			containsStr1: "[ WARN] log message",
			containsStr2: "",
		},
		{
			name:         "error",
			fn:           Error,
			containsStr1: "()] log message",
			containsStr2: "()] log message",
		},
		{
			name:         "fatal",
			fn:           Fatal,
			containsStr1: "()] log message",
			containsStr2: "()] log message",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer initManager()
			var buf1, buf2 bytes.Buffer
			assert.Nil(t, New(mode1, 2, bufConfig{
				buf: &buf1,
			}))
			assert.Nil(t, New(mode2, 2, bufConfig{
				buf: &buf2,
			}))
			assert.Equal(t, 2, mgr.len())

			tt.fn("log message")
			tt.fn("log message")
			Stop()

			assert.Contains(t, buf1.String(), tt.containsStr1)
			assert.Contains(t, buf2.String(), tt.containsStr2)
		})
	}
}
