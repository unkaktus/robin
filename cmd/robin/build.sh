#!/bin/bash

env GOOS=linux GOARCH=amd64 go build -o robin-linux-amd64
env GOOS=linux GOARCH=arm64 go build -o robin-linux-arm64
env GOOS=darwin GOARCH=amd64 go build -o robin-darwin-amd64
env GOOS=darwin GOARCH=arm64 go build -o robin-darwin-arm64
