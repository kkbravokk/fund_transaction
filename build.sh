#!/bin/sh
set -e
env

mkdir -p output && go build -o bin/transaction main.go
cp -rf bin default.yaml output