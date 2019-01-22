#!/bin/bash

env GOOS=linux go build -o server main.go
skaffold run