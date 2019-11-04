package clog

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ModeSlack(t *testing.T) {
	defer func() {
		Remove(ModeSlack)
	}()

	tests := []struct {
		name      string
		config    interface{}
		wantLevel Level
		wantErr   error
	}{
		{
			name: "valid config",
			config: SlackConfig{
				Level:      LevelInfo,
				BufferSize: 10,
				URL:        "https://slack.com",
				Colors:     slackColors,
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

func Test_buildSlackPayload(t *testing.T) {
	t.Run("default colors", func(t *testing.T) {
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
				payload, err := buildSlackPayload(slackColors, tt.msg)
				assert.Nil(t, err)
				assert.Equal(t, tt.want, payload)
			})
		}
	})

	t.Run("custom colors", func(t *testing.T) {
		colors := []string{"#1", "#2", "#3", "#4", "#5"}

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
				payload, err := buildSlackPayload(colors, tt.msg)
				assert.Nil(t, err)
				assert.Equal(t, tt.want, payload)
			})
		}
	})
}
