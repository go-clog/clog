package clog

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

const ModeSlack Mode = "slack"

type slackAttachment struct {
	Text  string `json:"text"`
	Color string `json:"color"`
}

type slackPayload struct {
	Attachments []slackAttachment `json:"attachments"`
}

var slackColors = []string{
	"",        // Trace
	"#3aa3e3", // Info
	"warning", // Warn
	"danger",  // Error
	"#ff0200", // Fatal
}

type SlackConfig struct {
	// Minimum level of messages to be processed.
	Level Level
	// Buffer size defines how many messages can be queued before hangs.
	BufferSize int64
	// Slack webhook URL.
	URL string
	// Colors for different levels, must have exact 5 elements in the order of
	// Trace, Info, Warn, Error, and Fatal.
	Colors []string
}

var _ Logger = (*slackLogger)(nil)

type slackLogger struct {
	level    Level
	msgChan  chan Messager
	doneChan chan struct{}

	url    string
	colors []string
}

func (_ *slackLogger) Mode() Mode {
	return ModeSlack
}

func (l *slackLogger) Level() Level {
	return l.level
}

func (l *slackLogger) Start(ctx context.Context) {
loop:
	for {
		select {
		case m := <-l.msgChan:
			l.postMessage(m)
		case <-ctx.Done():
			break loop
		}
	}

	for {
		if len(l.msgChan) == 0 {
			break
		}

		l.postMessage(<-l.msgChan)
	}
	l.doneChan <- struct{}{} // Notify the cleanup is done.
}

func buildSlackPayload(colors []string, m Messager) (string, error) {
	payload := slackPayload{
		Attachments: []slackAttachment{
			{
				Text:  m.String(),
				Color: colors[m.Level()],
			},
		},
	}
	p, err := json.Marshal(&payload)
	if err != nil {
		return "", err
	}
	return string(p), nil
}

func (l *slackLogger) postMessage(m Messager) {
	payload, err := buildSlackPayload(l.colors, m)
	if err != nil {
		fmt.Printf("slackLogger: error building payload: %v", err)
		return
	}

	resp, err := http.Post(l.url, "application/json", bytes.NewReader([]byte(payload)))
	if err != nil {
		fmt.Printf("slackLogger: error making HTTP request: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("slackLogger: error reading HTTP response body: %v", err)
			return
		}
		fmt.Printf("slackLogger: non-success response status code %d with body: %s", resp.StatusCode, data)
	}
}

func (l *slackLogger) Write(m Messager) {
	l.msgChan <- m
}

func (l *slackLogger) WaitForStop() {
	<-l.doneChan
}

func init() {
	NewRegister(ModeSlack, func(v interface{}) (Logger, error) {
		cfg, ok := v.(SlackConfig)
		if !ok {
			return nil, fmt.Errorf("invalid config object: want %T got %T", SlackConfig{}, v)
		}

		if cfg.URL == "" {
			return nil, errors.New("empty URL")
		}

		colors := slackColors
		if cfg.Colors != nil {
			if len(cfg.Colors) != 5 {
				return nil, fmt.Errorf("colors must have exact 5 elements, but got %d", len(cfg.Colors))
			}
			colors = cfg.Colors
		}

		return &slackLogger{
			level:    cfg.Level,
			msgChan:  make(chan Messager, cfg.BufferSize),
			doneChan: make(chan struct{}),
			url:      cfg.URL,
			colors:   colors,
		}, nil
	})
}
