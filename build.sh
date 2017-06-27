#!/bin/bash

GOOS=freebsd GOARCH=amd64 go build -o miniscope_freebsd_amd64 -x cmd/miniscope/main.go
GOOS=linux GOARCH=amd64 go build -o miniscope_linux_amd64 -x cmd/miniscope/main.go
GOOS=darwin GOARCH=amd64 go build -o miniscope_darwin_amd64 -x cmd/miniscope/main.go
GOOS=windows GOARCH=amd64 go build -o miniscope_windows_amd64 -x cmd/miniscope/main.go
