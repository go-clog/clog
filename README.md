# Clog [![Build Status](https://travis-ci.org/go-clog/clog.svg?branch=master)](https://travis-ci.org/go-clog/clog)

![](https://avatars1.githubusercontent.com/u/25576866?v=3&s=200)

Clog is a channel-based logging package for Go.

This package supports multiple logger adapters across different levels of logging. It uses Go's native channel feature to provide goroutine-safe mechanism on large concurrency.

## Installation

To use a tagged revision:

	go get gopkg.in/clog.v1

To use with latest changes:

	go get github.com/go-clog/clog
    
Please apply `-u` flag to update in the future.

### Testing

If you want to test on your machine, please apply `-t` flag:

	go get -t gopkg.in/clog.v1

Please apply `-u` flag to update in the future.

## Credits

- Avatar is a modified version based on [egonelbre/gophers' scientist](https://github.com/egonelbre/gophers/blob/master/vector/science/scientist.svg).

## License

This project is under Apache v2 License. See the [LICENSE](LICENSE) file for the full license text.