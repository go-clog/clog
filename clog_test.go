package clog

import (
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

type chanConfig struct {
	c chan string
}

var _ Logger = (*chanLogger)(nil)

type chanLogger struct {
	c chan string
	*noopLogger
}

func (l *chanLogger) Write(m Messager) error {
	fmt.Println(1, m)
	l.c <- m.String()
	fmt.Println(2, m)
	return nil
}

func Test_chanLogger(t *testing.T) {
	mode1 := Mode("mode1")
	level1 := LevelTrace
	NewRegister(mode1, func(v interface{}) (Logger, error) {
		cfg, ok := v.(chanConfig)
		if !ok {
			return nil, fmt.Errorf("invalid config object: want %T got %T", chanConfig{}, v)
		}
		return &chanLogger{
			c: cfg.c,
			noopLogger: &noopLogger{
				mode:  mode1,
				level: level1,
			},
		}, nil
	})

	mode2 := Mode("mode2")
	level2 := LevelError
	NewRegister(mode2, func(v interface{}) (Logger, error) {
		cfg, ok := v.(chanConfig)
		if !ok {
			return nil, fmt.Errorf("invalid config object: want %T got %T", &chanConfig{}, v)
		}
		return &chanLogger{
			c: cfg.c,
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

	c1 := make(chan string)
	c2 := make(chan string)
	assert.Nil(t, New(mode1, 1, chanConfig{
		c: c1,
	}))
	assert.Nil(t, New(mode2, 1, chanConfig{
		c: c2,
	}))

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
