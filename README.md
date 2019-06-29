# Snitch

Snitch is a binary that helps your TDD cycle or not so TDD by watching tests and implementations of Go files.
It works by scanning files, checking the modification date and when changed, giving feedback on your console.

It's usual in Go projects to keep the implementation and tests under the same package, so this binary follows this _convention_.

## Inspiration

It was a Friday afternoon and I was writing code, but had nothing to watch and report my tests while I changed code.

Inspired by [Guard](https://github.com/guard/guard), I decided to build this and thought more people could benefit from it.

## Requirements

Go 1.12+ installed. :heart:

## How to

### Compilation

`go build`

### Run

`./snitch --path <root-path> --time <time-in-seconds>`

The path and time params are both _optional_:

```
path: defaults to current dir
interval: defaults to 5s
```

If you have suggestions, doubts and bug reports, just let me know and let's improve it! I hope you enjoy it!
