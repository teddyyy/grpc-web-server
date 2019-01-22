#!/bin/bash

env GOOS=linux go build -o bin/server main.go
skaffold run
