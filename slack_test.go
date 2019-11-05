package clog

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ModeSlack(t *testing.T) {
	defer Remove(ModeSlack)

	tests := []struct {
		name      string
		config    interface{}
		wantLevel Level
		wantErr   error
	}{
		{
			name: "valid config",
			config: SlackConfig{
				Level:  LevelInfo,
				URL:    "https://slack.com",
				Colors: slackColors,
			},
			wantErr: nil,
		},
		{
			name:    "invalid config",
			config:  "random things",
			wantErr: errors.New("initialize logger: invalid config object: want clog.SlackConfig got string"),
		},
		{
			name:    "invalid URL",
			config:  SlackConfig{},
			wantErr: errors.New("initialize logger: empty URL"),
		},
		{
			name: "incorrect number of colors",
			config: SlackConfig{
				URL:    "https://slack.com",
				Colors: []string{},
			},
			wantErr: errors.New("initialize logger: colors must have exact 5 elements, but got 0"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantErr, New(ModeSlack, 10, tt.config))
		})
	}

	assert.Equal(t, 1, mgr.len())
	assert.Equal(t, ModeSlack, mgr.loggers[0].Mode())
	assert.Equal(t, LevelInfo, mgr.loggers[0].Level())
}

func Test_slackLogger_buildPayload(t *testing.T) {
	t.Run("default colors", func(t *testing.T) {
		l := &slackLogger{
			colors: slackColors,
		}

		tests := []struct {
			name string
			msg  *message
			want string
		}{
			{
				name: "trace",
				msg: &message{
					level: LevelTrace,
					body:  "test message",
				},
				want: `{"attachments":[{"text":"test message","color":""}]}`,
			},
			{
				name: "info",
				msg: &message{
					level: LevelInfo,
					body:  "test message",
				},
				want: `{"attachments":[{"text":"test message","color":"#3aa3e3"}]}`,
			},
			{
				name: "warn",
				msg: &message{
					level: LevelWarn,
					body:  "test message",
				},
				want: `{"attachments":[{"text":"test message","color":"warning"}]}`,
			},
			{
				name: "error",
				msg: &message{
					level: LevelError,
					body:  "test message",
				},
				want: `{"attachments":[{"text":"test message","color":"danger"}]}`,
			},
			{
				name: "fatal",
				msg: &message{
					level: LevelFatal,
					body:  "test message",
				},
				want: `{"attachments":[{"text":"test message","color":"#ff0200"}]}`,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				payload, err := l.buildPayload(tt.msg)
				assert.Nil(t, err)
				assert.Equal(t, tt.want, payload)
			})
		}
	})

	t.Run("custom colors", func(t *testing.T) {
		l := &slackLogger{
			colors: []string{"#1", "#2", "#3", "#4", "#5"},
		}

		tests := []struct {
			name string
			msg  *message
			want string
		}{
			{
				name: "trace",
				msg: &message{
					level: LevelTrace,
					body:  "test message",
				},
				want: `{"attachments":[{"text":"test message","color":"#1"}]}`,
			},
			{
				name: "info",
				msg: &message{
					level: LevelInfo,
					body:  "test message",
				},
				want: `{"attachments":[{"text":"test message","color":"#2"}]}`,
			},
			{
				name: "warn",
				msg: &message{
					level: LevelWarn,
					body:  "test message",
				},
				want: `{"attachments":[{"text":"test message","color":"#3"}]}`,
			},
			{
				name: "error",
				msg: &message{
					level: LevelError,
					body:  "test message",
				},
				want: `{"attachments":[{"text":"test message","color":"#4"}]}`,
			},
			{
				name: "fatal",
				msg: &message{
					level: LevelFatal,
					body:  "test message",
				},
				want: `{"attachments":[{"text":"test message","color":"#5"}]}`,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				payload, err := l.buildPayload(tt.msg)
				assert.Nil(t, err)
				assert.Equal(t, tt.want, payload)
			})
		}
	})
}

type roundTripFunc func(req *http.Request) *http.Response

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

func Test_slackLogger_postMessage(t *testing.T) {
	l := &slackLogger{
		client: &http.Client{
			Transport: roundTripFunc(func(req *http.Request) *http.Response {
				statusCode := 500
				respBody := ""
				switch req.URL.String() {
				case "https://slack.com/success":
					statusCode = 200
					respBody = `OK`
				case "https://slack.com/non-success-response-status-code":
					statusCode = 404
					respBody = `Page Not Found`
				}

				return &http.Response{
					StatusCode: statusCode,
					Body:       ioutil.NopCloser(bytes.NewBufferString(respBody)),
					Header:     make(http.Header),
				}
			}),
		},
	}

	tests := []struct {
		name    string
		url     string
		wantErr error
	}{
		{
			name:    "success",
			url:     "https://slack.com/success",
			wantErr: nil,
		},
		{
			name:    "non-success response status code",
			url:     "https://slack.com/non-success-response-status-code",
			wantErr: errors.New("non-success response status code 404 with body: Page Not Found"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l.url = tt.url
			assert.Equal(t, tt.wantErr, l.postMessage(bytes.NewReader([]byte("payload"))))
		})
	}
}
