#!/usr/bin/env fish

env CGO_LDFLAGS="-L"(pwd)"/cnutil -L"(pwd)"/hashing" go run main.go $argv
