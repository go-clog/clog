package clog

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const ModeFile Mode = "file"

const (
	simpleDateFormat = "2006-01-02"
	logPrefixLength  = len("2017/02/06 21:20:08 ")
)

// FileRotationConfig represents rotation related configurations for file mode logger.
// All the settings can take effect at the same time, remain zero values to disable them.
type FileRotationConfig struct {
	// Do rotation for output files.
	Rotate bool
	// Rotate on daily basis.
	Daily bool
	// Maximum size in bytes of file for a rotation.
	MaxSize int64
	// Maximum number of lines for a rotation.
	MaxLines int64
	// Maximum lifetime of a output file in days.
	MaxDays int64
}

type FileConfig struct {
	// Minimum level of messages to be processed.
	Level Level
	// Buffer size defines how many messages can be queued before hangs.
	BufferSize int64
	// File name to outout messages.
	Filename string
	// Rotation related configurations.
	FileRotationConfig
}

var _ Logger = (*fileLogger)(nil)

type fileLogger struct {
	// Indicates whether object is been used in standalone mode.
	standalone bool

	level    Level
	msgChan  chan Messager
	doneChan chan struct{}

	filename       string
	rotationConfig FileRotationConfig

	file         *os.File
	openDay      int
	currentSize  int64
	currentLines int64

	*log.Logger
}

func (_ *fileLogger) Mode() Mode {
	return ModeFile
}

func (l *fileLogger) Level() Level {
	return l.level
}

func (l *fileLogger) Start(ctx context.Context) {
loop:
	for {
		select {
		case m := <-l.msgChan:
			l.write(m)
		case <-ctx.Done():
			break loop
		}
	}

	for {
		if len(l.msgChan) == 0 {
			break
		}

		l.write(<-l.msgChan)
	}
	l.doneChan <- struct{}{} // Notify the cleanup is done.
}

var newLineBytes = []byte("\n")

func (l *fileLogger) initFile() (err error) {
	l.file, err = os.OpenFile(l.filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0660)
	if err != nil {
		return fmt.Errorf("OpenFile '%s': %v", l.filename, err)
	}

	l.Logger = log.New(l.file, "", log.Ldate|log.Ltime)
	return nil
}

// isExist checks whether a file or directory exists.
// It returns false when the file or directory does not exist.
func isExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

// rotateFilename returns next available rotate filename with given date.
func (l *fileLogger) rotateFilename(date string) string {
	filename := fmt.Sprintf("%s.%s", l.filename, date)
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

func (l *fileLogger) deleteOutdatedFiles() {
	_ = filepath.Walk(filepath.Dir(l.filename), func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() &&
			info.ModTime().Before(time.Now().Add(-24*time.Hour*time.Duration(l.rotationConfig.MaxDays))) &&
			strings.HasPrefix(filepath.Base(path), filepath.Base(l.filename)) {
			_ = os.Remove(path)
		}
		return nil
	})
}

func (l *fileLogger) initRotate() error {
	// Gather basic file info for rotation.
	fi, err := l.file.Stat()
	if err != nil {
		return fmt.Errorf("Stat: %v", err)
	}

	l.currentSize = fi.Size()

	// If there is any content in the file, count the number of lines.
	if l.rotationConfig.MaxLines > 0 && l.currentSize > 0 {
		data, err := ioutil.ReadFile(l.filename)
		if err != nil {
			return fmt.Errorf("ReadFile '%s': %v", l.filename, err)
		}

		l.currentLines = int64(bytes.Count(data, newLineBytes)) + 1
	}

	if l.rotationConfig.Daily {
		now := time.Now()
		l.openDay = now.Day()

		lastWriteTime := fi.ModTime()
		if lastWriteTime.Year() != now.Year() ||
			lastWriteTime.Month() != now.Month() ||
			lastWriteTime.Day() != now.Day() {

			if err = l.file.Close(); err != nil {
				return fmt.Errorf("Close: %v", err)
			}
			if err = os.Rename(l.filename, l.rotateFilename(lastWriteTime.Format(simpleDateFormat))); err != nil {
				return fmt.Errorf("Rename: %v", err)
			}

			if err = l.initFile(); err != nil {
				return fmt.Errorf("initFile: %v", err)
			}
		}
	}

	if l.rotationConfig.MaxDays > 0 {
		l.deleteOutdatedFiles()
	}
	return nil
}

func (l *fileLogger) write(m Messager) int {
	l.Logger.Print(m.String())

	bytesWrote := len(m.String())
	if !l.standalone {
		bytesWrote += logPrefixLength
	}
	if l.rotationConfig.Rotate {
		l.currentSize += int64(bytesWrote)
		l.currentLines += int64(strings.Count(m.String(), "\n")) + 1

		var (
			needsRotate = false
			rotateDate  time.Time
		)

		now := time.Now()
		if l.rotationConfig.Daily && now.Day() != l.openDay {
			needsRotate = true
			rotateDate = now.Add(-24 * time.Hour)

		} else if (l.rotationConfig.MaxSize > 0 && l.currentSize >= l.rotationConfig.MaxSize) ||
			(l.rotationConfig.MaxLines > 0 && l.currentLines >= l.rotationConfig.MaxLines) {
			needsRotate = true
			rotateDate = now
		}

		if needsRotate {
			_ = l.file.Close()
			if err := os.Rename(l.filename, l.rotateFilename(rotateDate.Format(simpleDateFormat))); err != nil {
				fmt.Printf("fileLogger: error renaming rotate file %q: %v\n", l.filename, err)
				return bytesWrote
			}

			if err := l.initFile(); err != nil {
				fmt.Printf("fileLogger: error initializing log file %q: %v\n", l.filename, err)
				return bytesWrote
			}

			l.openDay = now.Day()
			l.currentSize = 0
			l.currentLines = 0
		}
	}
	return bytesWrote
}

func (l *fileLogger) Write(m Messager) {
	l.msgChan <- m
}

func (l *fileLogger) WaitForStop() {
	<-l.doneChan
}

func (l *fileLogger) init() error {
	_ = os.MkdirAll(filepath.Dir(l.filename), os.ModePerm)
	if err := l.initFile(); err != nil {
		return fmt.Errorf("init file %q: %v", l.filename, err)
	}

	if l.rotationConfig.Rotate {
		if err := l.initRotate(); err != nil {
			return fmt.Errorf("init rotate: %v", err)
		}
	}
	return nil
}

func init() {
	NewRegister(ModeFile, func(v interface{}) (Logger, error) {
		cfg, ok := v.(FileConfig)
		if !ok {
			return nil, fmt.Errorf("invalid config object: want %T got %T", FileConfig{}, v)
		}

		l := &fileLogger{
			level:          cfg.Level,
			msgChan:        make(chan Messager, cfg.BufferSize),
			doneChan:       make(chan struct{}),
			filename:       cfg.Filename,
			rotationConfig: cfg.FileRotationConfig,
		}

		if err := l.init(); err != nil {
			return nil, err
		}

		return l, nil
	})
}

var _ io.Writer = (*fileWriter)(nil)

type fileWriter struct {
	*fileLogger
}

// NewFileWriter returns an io.Writer for synchronized file logger in standalone mode.
func NewFileWriter(filename string, cfg FileRotationConfig) (io.Writer, error) {
	f := &fileLogger{
		standalone:     true,
		filename:       filename,
		rotationConfig: cfg,
	}
	if err := f.init(); err != nil {
		return nil, err
	}

	return &fileWriter{f}, nil
}

// Write implements method of io.Writer interface.
func (w *fileWriter) Write(p []byte) (int, error) {
	return w.write(&message{
		body: string(p),
	}), nil
}
