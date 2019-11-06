package clog

import (
	"fmt"
	"log"

	"github.com/fatih/color"
)

// ModeConsole is used to indicate console logger.
const ModeConsole Mode = "console"

// Console color set for different levels.
var consoleColors = []func(a ...interface{}) string{
	color.New(color.FgBlue).SprintFunc(),   // Trace
	color.New(color.FgGreen).SprintFunc(),  // Info
	color.New(color.FgYellow).SprintFunc(), // Warn
	color.New(color.FgRed).SprintFunc(),    // Error
	color.New(color.FgHiRed).SprintFunc(),  // Fatal
}

// ConsoleConfig is the config object for the console logger.
type ConsoleConfig struct {
	// Minimum logging level of messages to be processed.
	Level Level
}

var _ Logger = (*consoleLogger)(nil)

type consoleLogger struct {
	level Level
	*log.Logger
}

func (*consoleLogger) Mode() Mode {
	return ModeConsole
}

func (l *consoleLogger) Level() Level {
	return l.level
}

func (l *consoleLogger) Write(m Messager) error {
	l.Print(consoleColors[m.Level()](m.String()))
	return nil
}

func init() {
	NewRegister(ModeConsole, func(v interface{}) (Logger, error) {
		if v == nil {
			v = ConsoleConfig{}
		}

		cfg, ok := v.(ConsoleConfig)
		if !ok {
			return nil, fmt.Errorf("invalid config object: want %T got %T", ConsoleConfig{}, v)
		}

		return &consoleLogger{
			level:  cfg.Level,
			Logger: log.New(color.Output, "", log.Ldate|log.Ltime),
		}, nil
	})
}
