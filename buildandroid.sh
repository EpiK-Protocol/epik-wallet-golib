#!/bin/bash



rm -rf ./dev/android/*
echo "building android..."
gomobile bind -target=android/arm64 -v -o ./dev/android/epik.aar ./epik ./hd 
echo "android build"
