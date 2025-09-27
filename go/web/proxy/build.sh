#!/usr/bin/env bash
set -e
docker build --no-cache --platform=linux/amd64 -t saichler/proxy:latest .
docker push saichler/proxy:latest
