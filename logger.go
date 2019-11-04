package clog

import (
	"context"
	"fmt"
	"sync/atomic"
)

// Logger is an interface for a logger with a specific mode and level.
type Logger interface {
	// Mode returns the mode of the logger.
	Mode() Mode
	// Level returns minimum level of the logger.
	Level() Level
	// Start starts the backgound process and a context for cancellation.
	Start(context.Context)
	// Write processes a Messager asynchronically.
	Write(Messager)
	// WaitForStop blocks until the logger is fully stopped.
	WaitForStop()
}

// Register is a factory function taht returns a new Logger.
// It accepts a configuration struct specifically for the Logger.
type Register func(interface{}) (Logger, error)

var registers = map[Mode]Register{}

// NewRegister adds a new factory function as a Register to the global map.
//
// This function is not concurrent safe.
func NewRegister(mode Mode, r Register) {
	if r == nil {
		panic("register is nil")
	}
	if registers[mode] != nil {
		panic(fmt.Sprintf("register with mode %q already exists", mode))
	}
	registers[mode] = r
}

type cancelableLogger struct {
	cancel context.CancelFunc
	Logger
}

type loggerManager struct {
	state   int64 // 0=stopping, 1=running
	ctx     context.Context
	cancel  context.CancelFunc
	loggers []*cancelableLogger
}

func (mgr *loggerManager) num() int {
	return len(mgr.loggers)
}

func (mgr *loggerManager) write(level Level, skip int, format string, v ...interface{}) {
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

func (mgr *loggerManager) stop() {
	// Make sure cancellation is only propagated once to prevent deadlock of WaitForStop.
	if !atomic.CompareAndSwapInt64(&mgr.state, 1, 0) {
		return
	}

	mgr.cancel()
	for _, l := range mgr.loggers {
		l.WaitForStop()
	}
}

var loggerMgr *loggerManager

func initLoggerManager() {
	ctx, cancel := context.WithCancel(context.Background())
	loggerMgr = &loggerManager{
		state:  1,
		ctx:    ctx,
		cancel: cancel,
	}
}

func init() {
	initLoggerManager()
}

// New initializes and appends a new Logger to the managed list.
// Calling this function multiple times will overwrite previous Logger with same mode.
//
// This function is not concurrent safe.
func New(mode Mode, cfg interface{}) error {
	r, ok := registers[mode]
	if !ok {
		return fmt.Errorf("no register for %q", mode)
	}

	logger, err := r(cfg)
	if err != nil {
		return fmt.Errorf("initialize logger: %v", err)
	}

	ctx, cancel := context.WithCancel(loggerMgr.ctx)
	cl := &cancelableLogger{
		cancel: cancel,
		Logger: logger,
	}

	// Check and replace previous logger.
	found := false
	for i, l := range loggerMgr.loggers {
		if l.Mode() == mode {
			found = true

			// Release previous logger.
			l.cancel()
			l.WaitForStop()

			loggerMgr.loggers[i] = cl
			break
		}
	}
	if !found {
		loggerMgr.loggers = append(loggerMgr.loggers, cl)
	}

	go logger.Start(ctx)
	return nil
}

// Remove removes a logger with given mode from the managed list.
//
// This function is not concurrent safe.
func Remove(mode Mode) {
	idx := -1
	for i, l := range loggerMgr.loggers {
		if l.Mode() == mode {
			idx = i
			l.cancel()
			l.WaitForStop()
		}
	}
	if idx == -1 {
		return
	}

	loggers := loggerMgr.loggers[:0]
	for i, l := range loggerMgr.loggers {
		if i == idx {
			continue
		}
		loggers = append(loggers, l)
	}
	loggerMgr.loggers = loggers
}
