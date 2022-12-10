#!/bin/sh

go get
echo "building linux/amd64"
GOOS=linux GOARCH=amd64 go build -o wow-profile-copy-linux-amd64 .
echo "building windows/amd64"
GOOS=windows GOARCH=amd64 go build -o wow-profile-copy-windows-amd64.exe .
echo "building darwin/amd64"
GOOS=darwin GOARCH=amd64 go build -o wow-profile-copy-darwin-amd64 .
echo "building darwin/arm64"
GOOS=darwin GOARCH=arm64 go build -o wow-profile-copy-darwin-aarch64 .