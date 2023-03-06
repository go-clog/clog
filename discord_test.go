package clog

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_discordLogger(t *testing.T) {
	testName := "Test_ModeDiscord"
	defer Remove(DefaultDiscordName)
	defer Remove(testName)

	tests := []struct {
		name      string
		mode      string
		config    interface{}
		wantLevel Level
		wantErr   error
	}{
		{
			name:    "nil config",
			mode:    DefaultDiscordName,
			wantErr: errors.New("initialize logger: config object with the type 'clog.DiscordConfig' not found"),
		},
		{
			name: "valid config",
			mode: DefaultDiscordName,
			config: DiscordConfig{
				Level:  LevelInfo,
				URL:    "https://discordapp.com",
				Titles: discordTitles,
				Colors: discordColors,
			},
			wantErr: nil,
		},
		{
			name: "custom name",
			mode: testName,
			config: DiscordConfig{
				URL: "https://discordapp.com",
			},
			wantErr: nil,
		},

		{
			name:    "invalid config",
			config:  "random things",
			wantErr: errors.New("initialize logger: config object with the type 'clog.DiscordConfig' not found"),
		},
		{
			name:    "invalid URL",
			config:  DiscordConfig{},
			wantErr: errors.New("initialize logger: empty URL"),
		},
		{
			name: "incorrect number of titles",
			config: DiscordConfig{
				URL:    "https://discordapp.com",
				Titles: []string{},
			},
			wantErr: errors.New("initialize logger: titles must have exact 5 elements, but got 0"),
		},
		{
			name: "incorrect number of colors",
			config: DiscordConfig{
				URL:    "https://discordapp.com",
				Colors: []int{},
			},
			wantErr: errors.New("initialize logger: colors must have exact 5 elements, but got 0"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantErr, NewDiscordWithName(tt.mode, 10, tt.config))
		})
	}

	assert.Equal(t, 2, mgr.len())
	assert.Equal(t, DefaultDiscordName, mgr.loggers[0].Name())
	assert.Equal(t, LevelInfo, mgr.loggers[0].Level())
	assert.Equal(t, testName, mgr.loggers[1].Name())
	assert.Equal(t, LevelDebug, mgr.loggers[1].Level())
}

func Test_discordLogger_buildPayload(t *testing.T) {
	t.Run("default titles and colors", func(t *testing.T) {
		l := &discordLogger{
			titles: discordTitles,
			colors: discordColors,
		}

		tests := []struct {
			name      string
			msg       *message
			wantTitle string
			wantDesc  string
			wantColor int
		}{
			{
				name: "trace",
				msg: &message{
					level: LevelDebug,
					body:  "[TRACE] test message",
				},
				wantTitle: discordTitles[0],
				wantDesc:  "test message",
				wantColor: discordColors[0],
			},
			{
				name: "info",
				msg: &message{
					level: LevelInfo,
					body:  "[ INFO] test message",
				},
				wantTitle: discordTitles[1],
				wantDesc:  "test message",
				wantColor: discordColors[1],
			},
			{
				name: "warn",
				msg: &message{
					level: LevelWarn,
					body:  "[ WARN] test message",
				},
				wantTitle: discordTitles[2],
				wantDesc:  "test message",
				wantColor: discordColors[2],
			},
			{
				name: "error",
				msg: &message{
					level: LevelError,
					body:  "[ERROR] test message",
				},
				wantTitle: discordTitles[3],
				wantDesc:  "test message",
				wantColor: discordColors[3],
			},
			{
				name: "fatal",
				msg: &message{
					level: LevelFatal,
					body:  "[FATAL] test message",
				},
				wantTitle: discordTitles[4],
				wantDesc:  "test message",
				wantColor: discordColors[4],
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				payload, err := l.buildPayload(tt.msg)
				assert.Nil(t, err)

				obj := &discordPayload{}
				assert.Nil(t, json.Unmarshal([]byte(payload), obj))
				assert.Len(t, obj.Embeds, 1)

				assert.Equal(t, tt.wantTitle, obj.Embeds[0].Title)
				assert.Equal(t, tt.wantDesc, obj.Embeds[0].Description)
				assert.NotEmpty(t, obj.Embeds[0].Timestamp)
				assert.Equal(t, tt.wantColor, obj.Embeds[0].Color)
			})
		}
	})

	t.Run("custom titles and colors", func(t *testing.T) {
		l := &discordLogger{
			titles: []string{"1", "2", "3", "4", "5"},
			colors: []int{1, 2, 3, 4, 5},
		}

		tests := []struct {
			name      string
			msg       *message
			wantTitle string
			wantDesc  string
			wantColor int
		}{
			{
				name: "trace",
				msg: &message{
					level: LevelDebug,
					body:  "[TRACE] test message",
				},
				wantTitle: l.titles[0],
				wantDesc:  "test message",
				wantColor: l.colors[0],
			},
			{
				name: "info",
				msg: &message{
					level: LevelInfo,
					body:  "[ INFO] test message",
				},
				wantTitle: l.titles[1],
				wantDesc:  "test message",
				wantColor: l.colors[1],
			},
			{
				name: "warn",
				msg: &message{
					level: LevelWarn,
					body:  "[ WARN] test message",
				},
				wantTitle: l.titles[2],
				wantDesc:  "test message",
				wantColor: l.colors[2],
			},
			{
				name: "error",
				msg: &message{
					level: LevelError,
					body:  "[ERROR] test message",
				},
				wantTitle: l.titles[3],
				wantDesc:  "test message",
				wantColor: l.colors[3],
			},
			{
				name: "fatal",
				msg: &message{
					level: LevelFatal,
					body:  "[FATAL] test message",
				},
				wantTitle: l.titles[4],
				wantDesc:  "test message",
				wantColor: l.colors[4],
			},

			{
				name: "trace",
				msg: &message{
					level: LevelDebug,
					body:  "test message",
				},
				wantTitle: l.titles[0],
				wantDesc:  "test message",
				wantColor: l.colors[0],
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				payload, err := l.buildPayload(tt.msg)
				assert.Nil(t, err)

				obj := &discordPayload{}
				assert.Nil(t, json.Unmarshal([]byte(payload), obj))
				assert.Len(t, obj.Embeds, 1)

				assert.Equal(t, tt.wantTitle, obj.Embeds[0].Title)
				assert.Equal(t, tt.wantDesc, obj.Embeds[0].Description)
				assert.NotEmpty(t, obj.Embeds[0].Timestamp)
				assert.Equal(t, tt.wantColor, obj.Embeds[0].Color)
			})
		}
	})
}

func Test_discordLogger_postMessage(t *testing.T) {
	l := &discordLogger{
		client: &http.Client{
			Transport: roundTripFunc(func(req *http.Request) *http.Response {
				statusCode := 500
				respBody := ""
				switch req.URL.String() {
				case "https://discordapp.com/success":
					statusCode = 200
					respBody = `OK`
				case "https://discordapp.com/non-success-response-status-code":
					statusCode = 404
					respBody = `Page Not Found`
				case "https://discordapp.com/retry-after":
					statusCode = 429
					respBody = `{"retry_after": 123456}`
				}

				return &http.Response{
					StatusCode: statusCode,
					Body:       io.NopCloser(bytes.NewBufferString(respBody)),
					Header:     make(http.Header),
				}
			}),
		},
	}

	tests := []struct {
		name      string
		url       string
		wantRetry int64
		wantErr   error
	}{
		{
			name:      "success",
			url:       "https://discordapp.com/success",
			wantRetry: -1,
			wantErr:   nil,
		},
		{
			name:      "non-success response status code",
			url:       "https://discordapp.com/non-success-response-status-code",
			wantRetry: -1,
			wantErr:   errors.New("non-success response status code 404 with body: Page Not Found"),
		},
		{
			name:      "retry after",
			url:       "https://discordapp.com/retry-after",
			wantRetry: 123456,
			wantErr:   nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l.url = tt.url
			retryAfter, err := l.postMessage(bytes.NewReader([]byte("payload")))
			assert.Equal(t, tt.wantRetry, retryAfter)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}
