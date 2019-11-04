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

func (l *bufLogger) Write(m Messager) {
	l.buf.WriteString(m.String())
}

func Test_memoryLogger(t *testing.T) {
	defer initLoggerManager()

	mode1 := Mode("mode1")
	level1 := LevelTrace
	var buf1 bytes.Buffer
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
	assert.Nil(t, New(mode1, bufConfig{
		buf: &buf1,
	}))

	mode2 := Mode("mode2")
	level2 := LevelError
	var buf2 bytes.Buffer
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
	assert.Nil(t, New(mode2, bufConfig{
		buf: &buf2,
	}))

	assert.Equal(t, 2, loggerMgr.num())

	t.Run("trace", func(t *testing.T) {
		defer func() {
			buf1.Reset()
			buf2.Reset()
		}()
		Trace("this is a trace log")

		assert.Equal(t, "[TRACE] this is a trace log", buf1.String())
		assert.Empty(t, buf2.String())
	})

	t.Run("info", func(t *testing.T) {
		defer func() {
			buf1.Reset()
			buf2.Reset()
		}()
		Info("this is a info log")

		assert.Equal(t, "[ INFO] this is a info log", buf1.String())
		assert.Empty(t, buf2.String())
	})

	t.Run("warn", func(t *testing.T) {
		defer func() {
			buf1.Reset()
			buf2.Reset()
		}()
		Warn("this is a warn log")

		assert.Equal(t, "[ WARN] this is a warn log", buf1.String())
		assert.Empty(t, buf2.String())
	})

	t.Run("error", func(t *testing.T) {
		defer func() {
			buf1.Reset()
			buf2.Reset()
		}()
		Error("this is an error log")

		assert.Contains(t, buf1.String(), "()] this is an error log")
		assert.Contains(t, buf2.String(), "()] this is an error log")
	})

	t.Run("fatal", func(t *testing.T) {
		defer func() {
			buf1.Reset()
			buf2.Reset()
		}()
		Fatal("this is a fatal log")

		assert.Contains(t, buf1.String(), "()] this is a fatal log")
		assert.Contains(t, buf2.String(), "()] this is a fatal log")
	})
}
