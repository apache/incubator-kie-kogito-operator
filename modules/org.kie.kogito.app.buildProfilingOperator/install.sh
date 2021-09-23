#!/bin/sh
set -e
cd /workspace

GOOS=linux GOARCH=amd64 go test -covermode=atomic -coverpkg=github.com/kiegroup/kogito-operator/... -c -tags testrunmain ./ -o manager