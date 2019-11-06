package clog

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ModeConsole(t *testing.T) {
	defer Remove(ModeConsole)

	tests := []struct {
		name      string
		config    interface{}
		wantLevel Level
		wantErr   error
	}{
		{
			name: "valid config",
			config: ConsoleConfig{
				Level: LevelInfo,
			},
			wantErr: nil,
		},
		{
			name:    "invalid config",
			config:  "random things",
			wantErr: errors.New("initialize logger: invalid config object: want clog.ConsoleConfig got string"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantErr, New(ModeConsole, 10, tt.config))
		})
	}

	assert.Equal(t, 1, mgr.len())
	assert.Equal(t, ModeConsole, mgr.loggers[0].Mode())
	assert.Equal(t, LevelInfo, mgr.loggers[0].Level())
}
