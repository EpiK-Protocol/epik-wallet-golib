#!/bin/bash

output=epik
prefix=EPIK_


rm -rf ./dev/ios/*


echo "building ios..."
gomobile bind -target=ios -o ./dev/ios/${output}.framework -prefix=${prefix} ./wallet
zip -q -r ./dev/ios/${output}.framework.zip ./dev/ios/${output}.framework
rm -rf ./dev/ios/${output}.framework
echo "ios build"