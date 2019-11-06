package clog

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ModeFile(t *testing.T) {
	defer Remove(ModeFile)

	tests := []struct {
		name      string
		config    interface{}
		wantLevel Level
		wantErr   error
	}{
		{
			name: "valid config",
			config: FileConfig{
				Level:    LevelInfo,
				Filename: filepath.Join(os.TempDir(), "Test_ModeFile"),
			},
			wantErr: nil,
		},
		{
			name:    "invalid config",
			config:  "random things",
			wantErr: errors.New("initialize logger: invalid config object: want clog.FileConfig got string"),
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
			assert.Equal(t, tt.wantErr, New(ModeFile, 10, tt.config))
		})
	}

	assert.Equal(t, 1, mgr.len())
	assert.Equal(t, ModeFile, mgr.loggers[0].Mode())
	assert.Equal(t, LevelInfo, mgr.loggers[0].Level())
}

func Test_rotateFilename(t *testing.T) {
	_ = os.MkdirAll("test", os.ModePerm)
	defer os.RemoveAll("test")

	filename := rotateFilename("test/Test_rotateFilename.log", "2017-03-05")
	assert.Equal(t, "test/Test_rotateFilename.log.2017-03-05", filename)
	assert.Nil(t, ioutil.WriteFile(filename, []byte(""), os.ModePerm))

	filename = rotateFilename("test/Test_rotateFilename.log", "2017-03-05")
	assert.Equal(t, "test/Test_rotateFilename.log.2017-03-05.001", filename)
	assert.Nil(t, ioutil.WriteFile(filename, []byte(""), os.ModePerm))

	filename = rotateFilename("test/Test_rotateFilename.log", "2017-03-05")
	assert.Equal(t, "test/Test_rotateFilename.log.2017-03-05.002", filename)
}
