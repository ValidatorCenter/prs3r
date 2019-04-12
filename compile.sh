#!/bin/sh

GOARCH=amd64 go build -ldflags "-s" -o prs3rd *.go