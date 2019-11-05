package clog

import (
	"context"
	"fmt"
	"log"
	"sync/atomic"

	"github.com/fatih/color"
)

// Logger is an interface for a logger with a specific mode and level.
type Logger interface {
	// Mode returns the mode of the logger.
	Mode() Mode
	// Level returns the minimum logging level of the logger.
	Level() Level
	// Write processes a Messager entry.
	Write(Messager) error
}

// Register is a factory function taht returns a new logger.
// It accepts a configuration struct specifically for the logger.
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
	cancel  context.CancelFunc
	msgChan chan Messager
	done    chan struct{}
	Logger
}

var errLogger = log.New(color.Output, "", log.Ldate|log.Ltime)
var errSprintf = color.New(color.FgRed).Sprintf

func (l *cancelableLogger) error(err error) {
	if err == nil {
		return
	}

	errLogger.Print(errSprintf("[clog] [%s]: %v", l.Mode(), err))
}

type manager struct {
	state   int64 // 0=stopping, 1=running
	ctx     context.Context
	cancel  context.CancelFunc
	loggers []*cancelableLogger
}

func (m *manager) len() int {
	return len(m.loggers)
}

func (m *manager) write(level Level, skip int, format string, v ...interface{}) {
	var msg *message
	for i := range mgr.loggers {
		if mgr.loggers[i].Level() > level {
			continue
		}

		if msg == nil {
			msg = newMessage(level, skip, format, v...)
		}

		mgr.loggers[i].msgChan <- msg
	}

	if msg == nil {
		errLogger.Print(errSprintf("[clog] no logger is available"))
	}
}

func (m *manager) stop() {
	// Make sure cancellation is only propagated once to prevent deadlock of WaitForStop.
	if !atomic.CompareAndSwapInt64(&m.state, 1, 0) {
		return
	}

	m.cancel()
	for _, l := range m.loggers {
		<-l.done
	}
}

var mgr *manager

func initManager() {
	ctx, cancel := context.WithCancel(context.Background())
	mgr = &manager{
		state:  1,
		ctx:    ctx,
		cancel: cancel,
	}
}

func init() {
	initManager()
}

// New initializes and appends a new logger to the managed list.
// Calling this function multiple times will overwrite previous logger with same mode.
//
// This function is not concurrent safe.
func New(mode Mode, bufferSize int64, cfg interface{}) error {
	r, ok := registers[mode]
	if !ok {
		return fmt.Errorf("no register for %q", mode)
	}

	l, err := r(cfg)
	if err != nil {
		return fmt.Errorf("initialize logger: %v", err)
	}

	ctx, cancel := context.WithCancel(mgr.ctx)
	cl := &cancelableLogger{
		cancel:  cancel,
		msgChan: make(chan Messager, bufferSize),
		done:    make(chan struct{}),
		Logger:  l,
	}

	// Check and replace previous logger
	found := false
	for i, l := range mgr.loggers {
		if l.Mode() == mode {
			found = true

			// Release previous logger
			l.cancel()
			<-l.done

			mgr.loggers[i] = cl
			break
		}
	}
	if !found {
		mgr.loggers = append(mgr.loggers, cl)
	}

	go func() {
	loop:
		for {
			select {
			case m := <-cl.msgChan:
				cl.error(cl.Write(m))
			case <-ctx.Done():
				break loop
			}
		}

		// Drain the msgChan at best effort
		for {
			if len(cl.msgChan) == 0 {
				break
			}

			cl.error(cl.Write(<-cl.msgChan))
		}

		// Notify the cleanup is done
		cl.done <- struct{}{}
	}()
	return nil
}

// Remove removes a logger with given mode from the managed list.
//
// This function is not concurrent safe.
func Remove(mode Mode) {
	loggers := mgr.loggers[:0]
	for _, l := range mgr.loggers {
		if l.Mode() == mode {
			go func(l *cancelableLogger) {
				l.cancel()
				<-l.done
			}(l)
			continue
		}
		loggers = append(loggers, l)
	}
	mgr.loggers = loggers
}
