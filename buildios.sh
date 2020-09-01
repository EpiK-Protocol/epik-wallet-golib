#!/bin/bash

output=epik
prefix=EPIK_


rm -rf ./dev/ios/*


echo "building ios..."
gomobile bind -target=ios -o ./dev/ios/${output}.framework -prefix=${prefix} -ldflags "-s -w"  ./epik ./hd
zip -q -r ./dev/ios/${output}.framework.zip ./dev/ios/${output}.framework
echo "ios build"