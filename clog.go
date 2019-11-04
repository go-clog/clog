// Clog is a channel-based logging package for Go.
package clog

import (
	"fmt"
	"os"
)

// Mode is the output source.
type Mode string

//const (
//	ModeConsole Mode = "console"
//	ModeFile    Mode = "file"
//	ModeSlack   Mode = "slack"
//	ModeDiscord Mode = "discord"
//)

// Level is the logging level.
type Level int

const (
	LevelTrace Level = iota
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
)

func (l Level) String() string {
	switch l {
	case LevelTrace:
		return "TRACE"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	case LevelFatal:
		return "FATAL"
	default:
		fmt.Printf("Unexpected value: %v\n", int(l))
		panic("unreachable")
	}
}

func write(level Level, skip int, format string, v ...interface{}) {
	var msg *message
	for i := range loggerMgr.loggers {
		if loggerMgr.loggers[i].Level() > level {
			continue
		}

		if msg == nil {
			msg = newMessage(level, skip, format, v...)
		}

		loggerMgr.loggers[i].Write(msg)
	}
}

func Trace(format string, v ...interface{}) {
	write(LevelTrace, 0, format, v...)
}

func Info(format string, v ...interface{}) {
	write(LevelInfo, 0, format, v...)
}

func Warn(format string, v ...interface{}) {
	write(LevelWarn, 0, format, v...)
}

func Error(format string, v ...interface{}) {
	ErrorDepth(4, format, v...)
}

func ErrorDepth(skip int, format string, v ...interface{}) {
	write(LevelError, skip, format, v...)
}

func Fatal(format string, v ...interface{}) {
	FatalDepth(4, format, v...)
}

var inTest = false

func FatalDepth(skip int, format string, v ...interface{}) {
	write(LevelFatal, skip, format, v...)
	Shutdown()

	if inTest {
		return
	}
	os.Exit(1)
}

func Shutdown() {
	loggerMgr.stop()
}
