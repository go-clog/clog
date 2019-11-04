package clog

import (
	"context"
	"fmt"
	"log"

	"github.com/fatih/color"
)

const ModeConsole Mode = "console"

// Console color set for different levels.
var consoleColors = []func(a ...interface{}) string{
	color.New(color.FgBlue).SprintFunc(),   // Trace
	color.New(color.FgGreen).SprintFunc(),  // Info
	color.New(color.FgYellow).SprintFunc(), // Warn
	color.New(color.FgRed).SprintFunc(),    // Error
	color.New(color.FgHiRed).SprintFunc(),  // Fatal
}

// ConsoleConfig is the config object for the "console" mode logger.
type ConsoleConfig struct {
	// Minimum level of messages to be processed.
	Level Level
	// Buffer size defines how many messages can be queued before hangs.
	BufferSize int64
}

var _ Logger = (*consoleLogger)(nil)

type consoleLogger struct {
	level    Level
	msgChan  chan Messager
	doneChan chan struct{}

	*log.Logger
}

func (_ *consoleLogger) Mode() Mode {
	return ModeConsole
}

func (l *consoleLogger) Level() Level {
	return l.level
}

func (l *consoleLogger) Start(ctx context.Context) {
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

func (l *consoleLogger) write(m Messager) {
	l.Print(consoleColors[m.Level()](m.String()))
}

func (l *consoleLogger) Write(m Messager) {
	l.msgChan <- m
}

func (l *consoleLogger) WaitForStop() {
	<-l.doneChan
}

func init() {
	NewRegister(ModeConsole, func(v interface{}) (Logger, error) {
		cfg, ok := v.(ConsoleConfig)
		if !ok {
			return nil, fmt.Errorf("invalid config object: want %T got %T", ConsoleConfig{}, v)
		}

		return &consoleLogger{
			level:    cfg.Level,
			msgChan:  make(chan Messager, cfg.BufferSize),
			doneChan: make(chan struct{}),
			Logger:   log.New(color.Output, "", log.Ldate|log.Ltime),
		}, nil
	})
}
