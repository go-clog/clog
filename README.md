# Clog 

[![Build Status](https://img.shields.io/travis/go-clog/clog/master.svg?style=for-the-badge&logo=travis)](https://travis-ci.org/go-clog/clog) [![codecov](https://img.shields.io/codecov/c/github/go-clog/clog/master?logo=codecov&style=for-the-badge)](https://codecov.io/gh/go-clog/clog) [![GoDoc](https://img.shields.io/badge/GoDoc-Reference-blue?style=for-the-badge&logo=go)](https://godoc.org/unknwon.dev/clog/v2) [![Sourcegraph](https://img.shields.io/badge/view%20on-Sourcegraph-brightgreen.svg?style=for-the-badge&logo=sourcegraph)](https://sourcegraph.com/github.com/go-clog/clog)

![](https://avatars1.githubusercontent.com/u/25576866?v=3&s=200)

Package clog is a channel-based logging package for Go.

This package supports multiple loggers across different levels of logging. It uses Go's native channel feature to provide goroutine-safe mechanism on large concurrency.

## Installation

To use a tagged revision:

	go get unknwon.dev/clog/v2
    
Please apply `-u` flag to update in the future.

### Testing

If you want to test on your machine, please apply `-t` flag:

	go get -t unknwon.dev/clog/v2

Please apply `-u` flag to update in the future.

## Getting Started

Clog currently has four builtin logger adapters: `console`, `file`, `slack` and `discord`.

It is extremely easy to create one with all default settings. Generally, you would want to create new logger inside `init` or `main` function.

```go
...

import (
	"fmt"
	"os"

	log "unknwon.dev/clog/v2"
)

func init() {
	// 0 means logging synchronously
	err := log.New(log.ModeConsole, 0, log.ConsoleConfig{})
	if err != nil {
		fmt.Printf("Fail to create new logger: %v\n", err)
		os.Exit(1)
	}

	log.Trace("Hello %s!", "Clog")
	// Output: Hello Clog!

	log.Info("Hello %s!", "Clog")
	log.Warn("Hello %s!", "Clog")
	...
	
	// Graceful stopping all loggers before exiting the program.
	log.Stop()
}

...
```

The above code is equivalent to the follow settings:

```go
...
	err := log.New(log.ModeConsole, 0, log.ConsoleConfig{
		Level:      log.LevelTrace, // Record all logs
	})
...
```

In production, you may want to make log less verbose and asynchronous:

```go
...
	// The buffer size mainly depends on how many logs will be produced at the same time,
	// 100 is a good default.
	err := log.New(log.ModeConsole, 100, log.ConsoleConfig{
		// Logs under Info level (in this case Trace) will be discarded.
		Level:      log.LevelInfo,
	})
...
```

Console logger comes with color output, but for non-colorable destination, the color output will be disabled automatically.

### Error Location

When using `log.Error` and `log.Fatal` functions, the caller location is printed along with the message. 

```go
...
	log.Error("So bad... %v", err)
	// Output: 2017/02/09 01:06:16 [ERROR] [...uban-builder/main.go:64 main()] ...
	log.Fatal("Boom! %v", err)
	// Output: 2017/02/09 01:06:16 [FATAL] [...uban-builder/main.go:64 main()] ...
...
```

Calling `log.Fatal` will exit the program.

If you want to have different skip depth than the default, you can use `log.ErrorDepth` or `log.FatalDepth`.

### Clean Exit

You should always call `log.Stop()` to wait until all messages are processed before program exits.

## File

File logger is more complex than console, and it has ability to rotate:

```go
...
	err := log.New(log.ModeFile, 100, log.FileConfig{
		Level:              log.LevelInfo, 
		Filename:           "clog.log",  
		FileRotationConfig: log.FileRotationConfig {
			Rotate: true,
			Daily:  true,
		},
	})
...
```

## Slack

Slack logger is also supported in a simple way:

```go
...
	err := log.New(log.ModeSlack, 100, log.SlackConfig{
		Level:              log.LevelInfo, 
		URL:                "https://url-to-slack-webhook",  
	})
...
```

This logger also works for [Discord Slack](https://discordapp.com/developers/docs/resources/webhook#execute-slackcompatible-webhook) endpoint.

## Discord

Discord logger is supported in rich format via [Embed Object](https://discordapp.com/developers/docs/resources/channel#embed-object):

```go
...
	err := log.New(log.ModeDiscord, 100, log.DiscordConfig{
		Level:              log.LevelInfo, 
		URL:                "https://url-to-discord-webhook",  
	})
...
```

This logger also retries automatically if hits rate limit after `retry_after`.

## Credits

- Avatar is a modified version based on [egonelbre/gophers' scientist](https://github.com/egonelbre/gophers/blob/master/vector/science/scientist.svg).

## License

This project is under MIT License. See the [LICENSE](LICENSE) file for the full license text.
