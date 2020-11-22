package clog

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func init() {
	isTestEnv = true
}

func TestLevel_String(t *testing.T) {
	invalidLevel := Level(-1)
	defer func() {
		assert.NotNil(t, recover())
	}()

	_ = invalidLevel.String()
}

type chanConfig struct {
	c chan string
}

var _ Logger = (*chanLogger)(nil)

type chanLogger struct {
	c chan string
	*noopLogger
}

func (l *chanLogger) Write(m Messager) error {
	l.c <- m.String()
	return nil
}

func chanLoggerIniter(name string, level Level) Initer {
	return func(_ string, vs ...interface{}) (Logger, error) {
		var cfg *chanConfig
		for i := range vs {
			switch v := vs[i].(type) {
			case chanConfig:
				cfg = &v
			}
		}

		if cfg == nil {
			return nil, fmt.Errorf("config object with the type '%T' not found", chanConfig{})
		}

		return &chanLogger{
			c: cfg.c,
			noopLogger: &noopLogger{
				name:  name,
				level: level,
			},
		}, nil
	}
}

func Test_chanLogger(t *testing.T) {
	test1 := "mode1"
	test1Initer := chanLoggerIniter(test1, LevelTrace)

	test2 := "mode2"
	test2Initer := chanLoggerIniter(test2, LevelError)

	c1 := make(chan string)
	c2 := make(chan string)

	defer Remove(test1)
	defer Remove(test2)
	assert.Nil(t, New(test1, test1Initer, 1, chanConfig{
		c: c1,
	}))
	assert.Nil(t, New(test2, test2Initer, 1, chanConfig{
		c: c2,
	}))

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
			assert.Equal(t, 2, mgr.len())

			tt.fn("log message")

			assert.Contains(t, <-c1, tt.containsStr1)

			if tt.containsStr2 != "" {
				assert.Contains(t, <-c2, tt.containsStr2)
			}
		})
	}
}

func Test_writeToNamedLogger(t *testing.T) {
	test1 := "alice"
	test1Initer := chanLoggerIniter(test1, LevelTrace)

	test2 := "bob"
	test2Initer := chanLoggerIniter(test2, LevelTrace)

	c1 := make(chan string)
	c2 := make(chan string)

	defer Remove(test1)
	defer Remove(test2)
	assert.Nil(t, New(test1, test1Initer, 1, chanConfig{
		c: c1,
	}))
	assert.Nil(t, New(test2, test2Initer, 1, chanConfig{
		c: c2,
	}))

	tests := []struct {
		name         string
		fn           func(string, string, ...interface{})
		containsStr1 string
		containsStr2 string
	}{
		{
			name:         "trace",
			fn:           TraceTo,
			containsStr1: "[TRACE] log message",
			containsStr2: "",
		},
		{
			name:         "info",
			fn:           InfoTo,
			containsStr1: "[ INFO] log message",
			containsStr2: "",
		},
		{
			name:         "warn",
			fn:           WarnTo,
			containsStr1: "[ WARN] log message",
			containsStr2: "",
		},
		{
			name:         "error",
			fn:           ErrorTo,
			containsStr1: "()] log message",
			containsStr2: "",
		},
		{
			name:         "fatal",
			fn:           FatalTo,
			containsStr1: "()] log message",
			containsStr2: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, 2, mgr.len())

			tt.fn(test1, "log message")

			assert.Contains(t, <-c1, tt.containsStr1)

			if tt.containsStr2 != "" {
				assert.Contains(t, <-c2, tt.containsStr2)
			}
		})
	}
}
