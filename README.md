# Snitch

Snitch is a binary that helps your TDD cycle (or not) by watching tests and implementations of Go files.
It works by scanning files, checking the modification date on save re-running your tests.

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

## Requirements

Go 1.12+ :heart:

The binary is _go-gettable_. Make sure you have `GOPATH` correctly set and added to the `$PATH`:

`go get github.com/axcdnt/snitch`

After _go-getting_ the binary, it will probably be available on your terminal.

## Run

`./snitch --path <root-path> --interval <time>`

The path and interval params are both _optional_:

```
path: defaults to current dir
interval: defaults to 5s
```

Interval command line argument accepts any [time.ParseDuration](https://golang.org/pkg/time/#ParseDuration)
"parseable" string, i.e. for faster scans you may use `--interval=500ms`.

During my tests I noticed that passing `path` for `go test` shows a peculiar behavior and cannot resolve it. An alternative use case is to always run `snitch` inside the project. Avoid the `--path` by now.

Feedback is welcome. I hope you enjoy it!
