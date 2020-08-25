# Snitch

[![contributions welcome](https://img.shields.io/badge/contributions-welcome-brightgreen.svg?style=flat)](https://github.com/fishybell/snitch/issues)
[![Build Status](https://travis-ci.org/fishybell/snitch.svg?branch=master)](https://travis-ci.org/fishybell/snitch)
[![Go Report Card](https://goreportcard.com/badge/github.com/fishybell/snitch)](https://goreportcard.com/report/github.com/fishybell/snitch)

<img src="https://github.com/fishybell/snitch/blob/master/logo.png" width="200">

Snitch is a binary that helps your TDD cycle (or not) by watching tests and implementations of Go files.
It works by scanning files, checking the modification date on save and re-running your tests.

It's usual in Go projects to keep the implementation and tests under the same package, so this binary follows this _convention_.

This tool focuses on Go developers. With a few LOCs we get interesting stuff.

## Inspiration

It was a Friday afternoon and I was writing code, but had nothing to watch and report my tests while I changed code.

Inspired by [Guard](https://github.com/guard/guard), I decided to build this and thought more people could benefit from it.

Forked from [snitch](https://github.com/axcdnt/snitch), but with quieter defaults.

## Features

- Automatically runs your tests
- Re-scan new files, so no need to restart
- Runs on a package basis
- Shows test coverage percentage
- Desktop notifications on macOS and Linux (via `notify-send`)

## Requirements

Go 1.12+

The binary is _go-gettable_. Make sure you have `GOPATH` correctly set and added to the `$PATH`:

`go get -u github.com/fishybell/snitch`

After _go-getting_ the binary, it will probably be available on your terminal.

## Run

```
▶ snitch --help
Usage of snitch:
  -d    Run with some debug output
  -f    [f]ull: Always run entire build
  -interval duration
        The interval (in seconds) for scanning files (default 1s)
  -m string
        The modules mode (passed to -mod= at test time) (default "mod")
  -n    [n]otify: Use system notifications
  -o    [o]nce: Only fail once, don't run subsequent tests
  -path string
        The root path to be watched (default "/Users/nbell/devel/rediq-deal-aggregator")
  -q    [q]uiet: Only print failing tests (use -q=false to be noisy again) (default true)
  -s    [s]mart: Run entire build when no test files are found
  -v    [v]ersion: Print the current version and exit
```

Feedback is welcome. I hope you enjoy it!
