#!/bin/bash
go get golang.org/x/mobile
go mod download golang.org/x/exp
rm -rf ./dev/android/*
echo "building android..."
gomobile bind -target=android/arm64 -o ./dev/android/epik.aar -ldflags "-s -w" -v ./epik ./hd
echo "android build"

output=epik
prefix=EPIK_

rm -rf ./dev/ios/*

echo "building ios..."
gomobile bind -target=ios -o ./dev/ios/${output}.framework -prefix=${prefix} -v ./epik ./hd
# zip -q -r ./dev/ios/${output}.framework.zip ./dev/ios/${output}.framework
echo "ios build"