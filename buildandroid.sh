#!/bin/bash


go get golang.org/x/mobile
go mod download golang.org/x/exp
rm -rf ./dev/android/*
echo "building android..."
gomobile bind -target=android/arm64 -v -o ./dev/android/epik.aar -ldflags "-s -w" ./epik ./hd 
echo "android build"
