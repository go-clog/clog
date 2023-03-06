package clog

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_fileLogger(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping testing on Windows")
	}

	testName := "Test_fileLogger"
	defer Remove(DefaultFileName)
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
			mode:    DefaultFileName,
			wantErr: nil,
		},
		{
			name: "valid config",
			mode: DefaultFileName,
			config: FileConfig{
				Level:    LevelInfo,
				Filename: filepath.Join(os.TempDir(), "Test_ModeFile"),
			},
			wantErr: nil,
		},
		{
			name:    "custom name",
			mode:    testName,
			wantErr: nil,
		},

		{
			name: "invalid filename",
			config: FileConfig{
				Level: LevelInfo,
			},
			wantErr: errors.New(`initialize logger: init file "": open file "": open : no such file or directory`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantErr, NewFileWithName(tt.mode, 10, tt.config))
		})
	}

	assert.Equal(t, 2, mgr.len())
	assert.Equal(t, DefaultFileName, mgr.loggers[0].Name())
	assert.Equal(t, LevelInfo, mgr.loggers[0].Level())
	assert.Equal(t, testName, mgr.loggers[1].Name())
	assert.Equal(t, LevelDebug, mgr.loggers[1].Level())
}

func Test_rotateFilename(t *testing.T) {
	_ = os.MkdirAll("test", os.ModePerm)
	defer os.RemoveAll("test")

	filename := rotateFilename("test/Test_rotateFilename.log", "2017-03-05")
	assert.Equal(t, "test/Test_rotateFilename.log.2017-03-05", filename)
	assert.Nil(t, os.WriteFile(filename, []byte(""), os.ModePerm))

	filename = rotateFilename("test/Test_rotateFilename.log", "2017-03-05")
	assert.Equal(t, "test/Test_rotateFilename.log.2017-03-05.001", filename)
	assert.Nil(t, os.WriteFile(filename, []byte(""), os.ModePerm))

	filename = rotateFilename("test/Test_rotateFilename.log", "2017-03-05")
	assert.Equal(t, "test/Test_rotateFilename.log.2017-03-05.002", filename)
}
