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
cp ./vendor/github.com/saichler/l8utils/go/utils/resources/build-test-security.sh .
chmod +x ./build-test-security.sh
rm -rf vendor
./build-test-security.sh
rm -rf ./build-test-security.sh

mkdir ./tmp
cd ./tmp
git clone https://github.com/saichler/l8test
cd ./l8test/go/infra/t_plugin/registry
./build.sh
mv *.so ../../../../../../tests/.
cd ../service
./build.sh
mv *.so ../../../../../../tests/.
cd ../../../../../../
rm -rf tmp

go mod vendor
rm -rf ./tests/test.*

# Run unit tests with coverage
go test -tags=unit -v -coverpkg=./web/... -coverprofile=cover.html ./... --failfast

#rm -rf ./tests/loader.so
#rm -rf ./tests/test.*

# Open the coverage report in a browser
go tool cover -html=cover.html
