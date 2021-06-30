#!/bin/sh
set -e
cd /workspace/api
go mod download
cd /workspace
go mod download