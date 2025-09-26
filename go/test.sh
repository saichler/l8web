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
go mod vendor

rm -rf ./tests/*.so

cp ../../l8secure/go/secure/provider/loader.so ./tests/loader.so

cd ../../l8test/go/infra/t_plugin/registry
./build.sh
cd ../service
./build.sh
cd ../../../../../l8web/go

cp ../../l8test/go/infra/t_plugin/registry/*.so ./tests/.
cp ../../l8test/go/infra/t_plugin/service/*.so ./tests/.

# Run unit tests with coverage
go test -tags=unit -v -coverpkg=./web/... -coverprofile=cover.html ./... --failfast

#rm -rf ./tests/loader.so
#rm -rf ./tests/test.*

# Open the coverage report in a browser
go tool cover -html=cover.html
