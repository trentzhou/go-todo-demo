#!/usr/bin/env bash

export CGO_ENABLED=0
go build go-todo-demo.go
docker build -t go-todo-demo .
