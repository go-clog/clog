package clog

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

const ModeDiscord Mode = "discord"

type (
	discordEmbed struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Timestamp   string `json:"timestamp"`
		Color       int    `json:"color"`
	}

	discordPayload struct {
		Username string          `json:"username,omitempty"`
		Embeds   []*discordEmbed `json:"embeds"`
	}
)

var (
	discordTitles = []string{
		"Trace",
		"Information",
		"Warning",
		"Error",
		"Fatal",
	}

	discordColors = []int{
		0,        // Trace
		3843043,  // Info
		16761600, // Warn
		13041721, // Error
		9440319,  // Fatal
	}
)

type DiscordConfig struct {
	// Minimum level of messages to be processed.
	Level Level
	// Discord webhook URL.
	URL string
	// Username to be shown for the message.
	// Leave empty to use default as set in the Discord.
	Username string
	// Title for different levels, must have exact 5 elements in the order of
	// Trace, Info, Warn, Error, and Fatal.
	Titles []string
	// Colors for different levels, must have exact 5 elements in the order of
	// Trace, Info, Warn, Error, and Fatal.
	Colors []int
}

var _ Logger = (*discordLogger)(nil)

type discordLogger struct {
	level    Level
	url      string
	username string
	titles   []string
	colors   []int
}

func (_ *discordLogger) Mode() Mode {
	return ModeDiscord
}

func (l *discordLogger) Level() Level {
	return l.level
}

func buildDiscordPayload(username string, titles []string, colors []int, m Messager) (string, error) {
	payload := discordPayload{
		Username: username,
		Embeds: []*discordEmbed{
			{
				Title:       titles[m.Level()],
				Description: m.String()[8:],
				Timestamp:   time.Now().Format(time.RFC3339),
				Color:       colors[m.Level()],
			},
		},
	}
	p, err := json.Marshal(&payload)
	if err != nil {
		return "", err
	}
	return string(p), nil
}

type discordRateLimitMsg struct {
	RetryAfter int64 `json:"retry_after"`
}

func (l *discordLogger) postMessage(r io.Reader) (int64, error) {
	resp, err := http.Post(l.url, "application/json", r)
	if err != nil {
		return -1, fmt.Errorf("HTTP request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 429 {
		rlMsg := &discordRateLimitMsg{}
		if err = json.NewDecoder(resp.Body).Decode(&rlMsg); err != nil {
			return -1, fmt.Errorf("decode rate limit message: %v", err)
		}

		return rlMsg.RetryAfter, nil
	} else if resp.StatusCode/100 != 2 {
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return -1, fmt.Errorf("read HTTP response body: %v", err)
		}
		return -1, fmt.Errorf("non-success response status code %d with body: %s", resp.StatusCode, data)
	}

	return -1, nil
}

func (l *discordLogger) Write(m Messager) error {
	payload, err := buildDiscordPayload(l.username, l.titles, l.colors, m)
	if err != nil {
		return fmt.Errorf("build payload: %v", err)
	}

	const retryTimes = 3
	// Due to discord limit, try at most x times with respect to "retry_after" parameter.
	for i := 1; i <= retryTimes; i++ {
		retryAfter, err := l.postMessage(bytes.NewReader([]byte(payload)))
		if err != nil {
			return fmt.Errorf("post message: %v", err)
		}

		if retryAfter > 0 {
			time.Sleep(time.Duration(retryAfter) * time.Millisecond)
			continue
		}

		return nil
	}

	return fmt.Errorf("gave up after %d retries", retryTimes)
}

func init() {
	NewRegister(ModeDiscord, func(v interface{}) (Logger, error) {
		cfg, ok := v.(DiscordConfig)
		if !ok {
			return nil, fmt.Errorf("invalid config object: want %T got %T", DiscordConfig{}, v)
		}

		if cfg.URL == "" {
			return nil, errors.New("empty URL")
		}

		titles := discordTitles
		if cfg.Titles != nil {
			if len(cfg.Titles) != 5 {
				return nil, fmt.Errorf("titles must have exact 5 elements, but got %d", len(cfg.Titles))
			}
			titles = cfg.Titles
		}

		colors := discordColors
		if cfg.Colors != nil {
			if len(cfg.Colors) != 5 {
				return nil, fmt.Errorf("colors must have exact 5 elements, but got %d", len(cfg.Colors))
			}
			colors = cfg.Colors
		}

		return &discordLogger{
			level:    cfg.Level,
			url:      cfg.URL,
			username: cfg.Username,
			titles:   titles,
			colors:   colors,
		}, nil
	})
}
