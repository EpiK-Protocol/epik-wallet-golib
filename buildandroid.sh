#!/bin/bash

output=epik
prefix=EPIK_


rm -rf ./dev/android/*
echo "building android..."
GOARCH=arm gomobile bind -target=android/arm64 -o ./dev/android/${output}.aar ./wallet ./api
echo "android build"
