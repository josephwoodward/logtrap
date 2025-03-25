#!/bin/bash

set -euo pipefail

## TODO: Improve this to make it all automated
prometheus --config.file ./internal/bench/prometheus.yml
go run internal/bench/server.go
go-wrk -c 2048 -d 120 http://localhost:8090/log-info
