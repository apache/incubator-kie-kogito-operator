#!/bin/sh
set -e
cd /workspace

go test -covermode=atomic -coverpkg=github.com/kiegroup/kogito-operator/... -c -tags testrunmain ./ -o manager
