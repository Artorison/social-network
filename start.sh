#!/bin/bash

set -e

docker compose up -d

sleep 10

go run ./cmd/redditclone/main.go