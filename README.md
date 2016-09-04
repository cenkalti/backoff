# Exponential Backoff [![GoDoc][godoc image]][godoc] [![Build Status][travis image]][travis] [![Coverage Status][coveralls image]][coveralls]

This is a Go port of the exponential backoff algorithm from [Google's HTTP Client Library for Java][google-http-java-client].

[Exponential backoff][exponential backoff wiki]
is an algorithm that uses feedback to multiplicatively decrease the rate of some process,
in order to gradually find an acceptable rate.
The retries exponentially increase and stop increasing when a certain threshold is met.

## Usage

See https://godoc.org/github.com/cenk/backoff#pkg-examples

[godoc]: https://godoc.org/github.com/cenk/backoff
[godoc image]: https://godoc.org/github.com/cenk/backoff?status.png
[travis]: https://travis-ci.org/cenk/backoff
[travis image]: https://travis-ci.org/cenk/backoff.png?branch=master
[coveralls]: https://coveralls.io/github/cenk/backoff?branch=master
[coveralls image]: https://coveralls.io/repos/github/cenk/backoff/badge.svg?branch=master

[google-http-java-client]: https://github.com/google/google-http-java-client
[exponential backoff wiki]: http://en.wikipedia.org/wiki/Exponential_backoff

[advanced example]: https://godoc.org/github.com/cenk/backoff#example_
