package clog

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

// ModeSlack is used to indicate Slack logger.
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

// SlackConfig is the config object for the Slack logger.
type SlackConfig struct {
	// Minimum logging level of messages to be processed.
	Level Level
	// Slack webhook URL.
	URL string
	// Colors for different levels, must have exact 5 elements in the order of
	// Trace, Info, Warn, Error, and Fatal.
	Colors []string
}

var _ Logger = (*slackLogger)(nil)

type slackLogger struct {
	level  Level
	url    string
	colors []string

	client *http.Client
}

func (*slackLogger) Mode() Mode {
	return ModeSlack
}

func (l *slackLogger) Level() Level {
	return l.level
}

func (l *slackLogger) buildPayload(m Messager) (string, error) {
	payload := slackPayload{
		Attachments: []slackAttachment{
			{
				Text:  m.String(),
				Color: l.colors[m.Level()],
			},
		},
	}
	p, err := json.Marshal(&payload)
	if err != nil {
		return "", err
	}
	return string(p), nil
}

func (l *slackLogger) postMessage(r io.Reader) error {
	resp, err := l.client.Post(l.url, "application/json", r)
	if err != nil {
		return fmt.Errorf("HTTP request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("read HTTP response body: %v", err)
		}
		return fmt.Errorf("non-success response status code %d with body: %s", resp.StatusCode, data)
	}
	return nil
}

func (l *slackLogger) Write(m Messager) error {
	payload, err := l.buildPayload(m)
	if err != nil {
		return fmt.Errorf("build payload: %v", err)
	}

	err = l.postMessage(bytes.NewReader([]byte(payload)))
	if err != nil {
		return fmt.Errorf("post message: %v", err)
	}
	return nil
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
			level:  cfg.Level,
			url:    cfg.URL,
			colors: colors,
			client: http.DefaultClient,
		}, nil
	})
}
