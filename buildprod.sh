#!/bin/bash

output=option
prefix=BFSS_


rm -rf ./prod/*
echo "building android..."
gomobile bind -target=android -o ./dev/${output}.aar -ldflags "-s -w" ./option ./wallet
echo "android build"
echo "building ios..."
gomobile bind -target=ios -o ./dev/${output}.framework -prefix=${prefix} -ldflags "-s -w" ./option ./wallet
zip -q -r ./dev/${output}.framework.zip ./dev/${output}.framework
rm -rf ./dev/${output}.framework
echo "ios build"
