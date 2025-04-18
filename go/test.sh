#!/usr/bin/env bash

# Fail on errors and don't open cover file
set -e
# clean up
rm -rf go.mod
rm -rf go.sum
rm -rf vendor

# fetch dependencies
go mod init
GOPROXY=direct GOPRIVATE=github.com go mod tidy
./build-security.sh
go mod vendor

# Run unit tests with coverage
go test -tags=unit -v -coverpkg=./web/... -coverprofile=cover.html ./... --failfast

# Open the coverage report in a browser
go tool cover -html=cover.html
