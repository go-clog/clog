package clog

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_consoleLogger(t *testing.T) {
	testName := "Test_consoleLogger"
	defer Remove(DefaultConsoleName)
	defer Remove(testName)

	tests := []struct {
		name      string
		mode      string
		config    interface{}
		wantLevel Level
		wantErr   error
	}{
		{
			name:    "nil config",
			mode:    DefaultConsoleName,
			wantErr: nil,
		},
		{
			name: "valid config",
			mode: DefaultConsoleName,
			config: ConsoleConfig{
				Level: LevelInfo,
			},
			wantErr: nil,
		},
		{
			name:    "custom name",
			mode:    testName,
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantErr, NewConsoleWithName(tt.mode, 10, tt.config))
		})
	}

	assert.Equal(t, 2, mgr.len())
	assert.Equal(t, DefaultConsoleName, mgr.loggers[0].Name())
	assert.Equal(t, LevelInfo, mgr.loggers[0].Level())
	assert.Equal(t, testName, mgr.loggers[1].Name())
	assert.Equal(t, LevelDebug, mgr.loggers[1].Level())
}
