#!/usr/bin/env fish

env GORACE="log_path=race.log" env CGO_LDFLAGS="-L"(pwd)"/cnutil -L"(pwd)"/hashing" go run -race main.go $argv
