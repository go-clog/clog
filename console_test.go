package clog

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ModeConsole(t *testing.T) {
	defer func() {
		Remove(ModeConsole)
	}()

	tests := []struct {
		name      string
		config    interface{}
		wantLevel Level
		wantErr   error
	}{
		{
			name: "valid config",
			config: ConsoleConfig{
				Level:      LevelInfo,
				BufferSize: 10,
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
			assert.Equal(t, tt.wantErr, New(ModeConsole, tt.config))
		})
	}

	assert.Equal(t, 1, loggerMgr.num())
	assert.Equal(t, LevelInfo, loggerMgr.loggers[0].Level())
}
