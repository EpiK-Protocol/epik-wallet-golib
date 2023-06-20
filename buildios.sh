#!/bin/bash

output=epik
prefix=EPIK_


rm -rf ./dev/ios/*

go get golang.org/x/mobile
echo "building ios..."
gomobile bind -target=ios -o ./dev/ios/${output}.xcframework -prefix=${prefix} -v -ldflags "-s -w" ./epik ./hd
# zip -q -r ./dev/ios/${output}.framework.zip ./dev/ios/${output}.framework
echo "ios build"