# Snitch

[![contributions welcome](https://img.shields.io/badge/contributions-welcome-brightgreen.svg?style=flat)](https://github.com/axcdnt/snitch/issues)
[![Build Status](https://travis-ci.org/axcdnt/snitch.svg?branch=master)](https://travis-ci.org/axcdnt/snitch)
[![Go Report Card](https://goreportcard.com/badge/github.com/axcdnt/snitch)](https://goreportcard.com/report/github.com/axcdnt/snitch)

<img src="https://github.com/axcdnt/snitch/blob/master/logo.png" width="200">

Snitch is a binary that helps your TDD cycle (or not) by watching tests and implementations of Go files.
It works by scanning files, checking the modification date on save and re-running your tests.

It's usual in Go projects to keep the implementation and tests under the same package, so this binary follows this _convention_.

This tool focuses on Go developers. With a few LOCs we get interesting stuff.

## Inspiration

It was a Friday afternoon and I was writing code, but had nothing to watch and report my tests while I changed code.

Inspired by [Guard](https://github.com/guard/guard), I decided to build this and thought more people could benefit from it.

## Features

- Automatically runs your tests
- Re-scan new files, so no need to restart
- Runs on a package basis
- Shows test coverage percentage
- Desktop notifications on macOS and Linux (via `notify-send`)

## Requirements

Go 1.12+ :heart:

The binary is _go-gettable_. Make sure you have `GOPATH` correctly set and added to the `$PATH`:

`go get github.com/axcdnt/snitch`

After _go-getting_ the binary, it will probably be available on your terminal.

## Run

```
▶ snitch --help
Usage of snitch:
  -interval duration
    	the interval (in seconds) for scanning files (default 1s)
  -path string
    	the root path to be watched (default "<current-dir>")
  -v    Print the current version and exit
```

Feedback is welcome. I hope you enjoy it!
