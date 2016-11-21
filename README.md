# Cscope HTML viewer

go-cscope is a code browser, powered by cscope and Golang. This is inspired by code search and cross reference engine like OpenGrok.

## Features

- HTML interface for cscope
- Colors!
- Easy to use and install
- No dependencies or external files needed -- apart from the single binary
- Uses your existing cscope database

## Installation

### Prebuilt binaries

Not available yet.

### Building from source

Make sure you have Go 1.5 or greater installed.

```
go install github.com/edwardchoh/go-cscope
```

## Usage

Execute go-cscope from a directory with a cscope.out file, or specify it on the command line:

```
go-cscope --path $HOME/src/project/
```