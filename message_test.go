package clog

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_newMessage(t *testing.T) {
	t.Run("No skip", func(t *testing.T) {
		tests := []struct {
			name   string
			level  Level
			format string
			v      []interface{}
			want   string
		}{
			{
				name:   "trace",
				level:  LevelTrace,
				format: "a trace log: %v",
				v:      []interface{}{"value"},
				want:   "[TRACE] a trace log: value",
			},
			{
				name:   "info",
				level:  LevelInfo,
				format: "a info log: %v",
				v:      []interface{}{"value"},
				want:   "[ INFO] a info log: value",
			},
			{
				name:   "warn",
				level:  LevelWarn,
				format: "a warn log: %v",
				v:      []interface{}{"value"},
				want:   "[ WARN] a warn log: value",
			},
			{
				name:   "error",
				level:  LevelError,
				format: "an error log: %v",
				v:      []interface{}{"value"},
				want:   "[ERROR] an error log: value",
			},
			{
				name:   "fatal",
				level:  LevelFatal,
				format: "a fatal log: %v",
				v:      []interface{}{"value"},
				want:   "[FATAL] a fatal log: value",
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				m := newMessage(tt.level, 0, tt.format, tt.v...)
				assert.Equal(t, tt.level, m.Level())
				assert.Equal(t, tt.want, m.String())
			})
		}
	})

	t.Run("Has skip", func(t *testing.T) {
		tests := []struct {
			name     string
			level    Level
			format   string
			v        []interface{}
			contains string
		}{
			{
				name:     "trace",
				level:    LevelTrace,
				format:   "a trace log: %v",
				v:        []interface{}{"value"},
				contains: "[TRACE] a trace log: value",
			},
			{
				name:     "info",
				level:    LevelInfo,
				format:   "a info log: %v",
				v:        []interface{}{"value"},
				contains: "[ INFO] a info log: value",
			},
			{
				name:     "warn",
				level:    LevelWarn,
				format:   "a warn log: %v",
				v:        []interface{}{"value"},
				contains: "[ WARN] a warn log: value",
			},
			{
				name:     "error",
				level:    LevelError,
				format:   "an error log: %v",
				v:        []interface{}{"value"},
				contains: "an error log: value",
			},
			{
				name:     "fatal",
				level:    LevelFatal,
				format:   "a fatal log: %v",
				v:        []interface{}{"value"},
				contains: "()] a fatal log: value",
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				m := newMessage(tt.level, 1, tt.format, tt.v...)
				assert.Equal(t, tt.level, m.Level())
				assert.Contains(t, m.String(), tt.contains)
			})
		}
	})
}
