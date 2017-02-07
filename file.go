// Copyright 2017 Unknwon
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package clog

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"
)

const (
	FILE                     MODE = "file"
	LAST_WRITE_FILENAME           = ".clog-lastwrite"
	LAST_WRITE_FORMAT             = "2006-01-02 MST"
	LAST_WRITE_FORMAT_SIMPLE      = "2006-01-02"
)

// FileRotationConfig represents rotation related configurations for file mode logger.
// All the settings can take effect at the same time, remain zero values to disable them.
type FileRotationConfig struct {
	// Do rotation for output files.
	Rotate bool
	// Rotate on daily basis.
	Daily bool
	// Maximum size of file for a rotation.
	MaxSize int64
	// Maximum number of lines for a rotation.
	MaxLines int64
	// Maximum lifetime of a output file in days.
	MaxDays int64
}

type FileConfig struct {
	// Minimum level of messages to be processed.
	Level LEVEL
	// Buffer size defines how many messages can be queued before hangs.
	BufferSize int64
	// File name to outout messages.
	Filename string
	// Rotation related configurations.
	FileRotationConfig
}

type file struct {
	*log.Logger
	Adapter

	file         *os.File
	filename     string
	openDay      int
	currentSize  int64
	currentLines int64
	rotate       FileRotationConfig
}

func newFile() Logger {
	return &file{
		Adapter: Adapter{
			quitChan: make(chan struct{}),
		},
	}
}

func (f *file) Level() LEVEL { return f.level }

var newLineBytes = []byte("\n")

// isExist checks whether a file or directory exists.
// It returns false when the file or directory does not exist.
func isExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

func (f *file) initFile() (err error) {
	f.file, err = os.OpenFile(f.filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0660)
	if err != nil {
		return fmt.Errorf("OpenFile '%s': %v", f.filename, err)
	}

	f.Logger = log.New(f.file, "", log.Ldate|log.Ltime)
	return nil
}

// rotateFilename returns next available rotate filename with given date.
func (f *file) rotateFilename(date string) string {
	filename := fmt.Sprintf("%s.%s", f.filename, date)
	if !isExist(filename) {
		return filename
	}

	format := filename + ".%03d"
	for i := 1; i < 1000; i++ {
		filename := fmt.Sprintf(format, i)
		if !isExist(filename) {
			return filename
		}
	}

	panic("too many log files for yesterday")
}

func (f *file) Init(v interface{}) (err error) {
	cfg, ok := v.(FileConfig)
	if !ok {
		return ErrConfigObject{"FileConfig", v}
	}

	if !isValidLevel(cfg.Level) {
		return ErrInvalidLevel{}
	}
	f.level = cfg.Level

	f.filename = cfg.Filename
	if err = f.initFile(); err != nil {
		return fmt.Errorf("initFile: %v", err)
	}

	// Gather basic file info for rotation.
	f.rotate = cfg.FileRotationConfig
	if f.rotate.Rotate {
		fi, err := f.file.Stat()
		if err != nil {
			return fmt.Errorf("Stat: %v", err)
		}

		f.currentSize = fi.Size()

		// If there is any content in the file, count the number of lines.
		if f.rotate.MaxLines > 0 && f.currentSize > 0 {
			data, err := ioutil.ReadFile(f.filename)
			if err != nil {
				return fmt.Errorf("ReadFile '%s': %v", f.filename, err)
			}

			f.currentLines = int64(bytes.Count(data, newLineBytes)) + 1
		}

		// Setup timer for next rotation of day.
		if f.rotate.Daily {
			now := time.Now()
			f.openDay = now.Day()

			// Load last write time if present.
			lastWritePath := filepath.Join(filepath.Dir(f.filename), LAST_WRITE_FILENAME)
			if isExist(lastWritePath) {
				data, err := ioutil.ReadFile(lastWritePath)
				if err != nil {
					return fmt.Errorf("ReadFile '%s': %v", lastWritePath, err)
				}
				lastWriteTime, err := time.Parse(LAST_WRITE_FORMAT, string(data))
				if err == nil &&
					(lastWriteTime.Year() != now.Year() ||
						lastWriteTime.Month() != now.Month() ||
						lastWriteTime.Day() != now.Day()) {

					f.file.Close()
					if err = os.Rename(f.filename, f.rotateFilename(lastWriteTime.Format(LAST_WRITE_FORMAT_SIMPLE))); err != nil {
						return fmt.Errorf("Rename: %v", err)
					}

					if err = f.initFile(); err != nil {
						return fmt.Errorf("initFile: %v", err)
					}
				}
			}

			// Save new write time.
			if err = ioutil.WriteFile(lastWritePath, []byte(now.Format(LAST_WRITE_FORMAT)), 0660); err != nil {
				return fmt.Errorf("WriteFile '%s': %v", lastWritePath, err)
			}
		}

		// Delete outdated files.
		if f.rotate.MaxDays > 0 {
		}
	}

	f.msgChan = make(chan *Message, cfg.BufferSize)
	return nil
}

func (f *file) ExchangeChans(errorChan chan<- error) chan *Message {
	f.errorChan = errorChan
	return f.msgChan
}

func (f *file) write(msg *Message) {
	f.Logger.Print(msg.Body)
}

func (f *file) Start() {
LOOP:
	for {
		select {
		case msg := <-f.msgChan:
			f.write(msg)
		case <-f.quitChan:
			break LOOP
		}
	}

	for {
		if len(f.msgChan) == 0 {
			break
		}

		f.write(<-f.msgChan)
	}
	f.quitChan <- struct{}{} // Notify the cleanup is done.
}

func (f *file) Destroy() {
	f.quitChan <- struct{}{}
	<-f.quitChan

	close(f.msgChan)
	close(f.quitChan)

	f.file.Close()
}

func init() {
	Register(FILE, newFile)
}
