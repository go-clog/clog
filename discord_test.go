package clog

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ModeDiscord(t *testing.T) {
	defer func() {
		Remove(ModeDiscord)
	}()

	tests := []struct {
		name      string
		config    interface{}
		wantLevel Level
		wantErr   error
	}{
		{
			name: "valid config",
			config: DiscordConfig{
				Level:      LevelInfo,
				BufferSize: 10,
				URL:        "https://discordapp.com",
				Titles:     discordTitles,
				Colors:     discordColors,
			},
			wantErr: nil,
		},
		{
			name:    "invalid config",
			config:  "random things",
			wantErr: errors.New("initialize logger: invalid config object: want clog.DiscordConfig got string"),
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
			assert.Equal(t, tt.wantErr, New(ModeDiscord, tt.config))
		})
	}

	assert.Equal(t, 1, loggerMgr.num())
	assert.Equal(t, ModeDiscord, loggerMgr.loggers[0].Mode())
	assert.Equal(t, LevelInfo, loggerMgr.loggers[0].Level())
}

func Test_buildDiscordPayload(t *testing.T) {
	t.Run("default titles and colors", func(t *testing.T) {
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
					level: LevelTrace,
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
				payload, err := buildDiscordPayload("", discordTitles, discordColors, tt.msg)
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
		titles := []string{"1", "2", "3", "4", "5"}
		colors := []int{1, 2, 3, 4, 5}

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
					level: LevelTrace,
					body:  "[TRACE] test message",
				},
				wantTitle: titles[0],
				wantDesc:  "test message",
				wantColor: colors[0],
			},
			{
				name: "info",
				msg: &message{
					level: LevelInfo,
					body:  "[ INFO] test message",
				},
				wantTitle: titles[1],
				wantDesc:  "test message",
				wantColor: colors[1],
			},
			{
				name: "warn",
				msg: &message{
					level: LevelWarn,
					body:  "[ WARN] test message",
				},
				wantTitle: titles[2],
				wantDesc:  "test message",
				wantColor: colors[2],
			},
			{
				name: "error",
				msg: &message{
					level: LevelError,
					body:  "[ERROR] test message",
				},
				wantTitle: titles[3],
				wantDesc:  "test message",
				wantColor: colors[3],
			},
			{
				name: "fatal",
				msg: &message{
					level: LevelFatal,
					body:  "[FATAL] test message",
				},
				wantTitle: titles[4],
				wantDesc:  "test message",
				wantColor: colors[4],
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				payload, err := buildDiscordPayload("", titles, colors, tt.msg)
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
