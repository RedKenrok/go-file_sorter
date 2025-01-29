#!/usr/bin/env bash

if [ -z "$1" ]; then
    DATE=$(date -u '+%Y-%m-%d.1')
    read -p "Enter version number (e.g. ${DATE}): " VERSION
else
    VERSION=$1
fi
if ! [[ $VERSION =~ ^[0-9]+-[0-9]+-[0-9]+\.[0-9]+$ ]]; then
    echo "Error: Version should be in format Y-m-d.v (e.g. 2025-01-25.1)"
    exit 1
fi

COMMIT=$(git rev-parse --short HEAD)
BUILD_DATE=$(date -u '+%Y-%m-%d %H:%M:%S')

PACKAGE_NAME="file_sorter"

#the full list of the platforms: https://golang.org/doc/install/source#environment
platforms=(
  "darwin/amd64"
  "darwin/arm64"

  "freebsd/386"
  "freebsd/amd64"
  "freebsd/arm"

  "linux/386"
  "linux/amd64"
  "linux/arm"
  "linux/arm64"

  "netbsd/386"
  "netbsd/amd64"
  "netbsd/arm"

  "openbsd/386"
  "openbsd/amd64"
  "openbsd/arm"
  "openbsd/arm64"

  "windows/386"
  "windows/amd64"
  "windows/arm"
  "windows/arm64"
)

printf "Building: ${PACKAGE_NAME}\nVersion: ${VERSION}\nCommit: ${COMMIT}\nBuild date: ${BUILD_DATE}\n\n"

for platform in "${platforms[@]}"
do
  platform_split=(${platform//\// })
  GOOS=${platform_split[0]}
  GOARCH=${platform_split[1]}
  output_name=$PACKAGE_NAME'-'$GOOS'-'$GOARCH
  if [ $GOOS = "windows" ]; then
    output_name+='.exe'
  fi

  echo "Started building $output_name."
  env CGO_ENABLED=0 GOOS=$GOOS GOARCH=$GOARCH go build -ldflags="-X 'main.version=${VERSION}' -X 'main.commit=${COMMIT}' -X 'main.buildDate=${BUILD_DATE}' -s -w" -o="./build/$output_name" ./main.go
  if [ $? -ne 0 ]; then
    echo 'An error has occurred! Aborting the script execution...'
    exit 1
  fi
  echo "Finished building $output_name."

  if [[ $GOOS == "linux" || $GOOS == "windows" ]]; then
    if [[ $platform == "windows/arm" || $platform == "windows/arm64" ]]; then
      echo "Skipping compression of $output_name."
    else
      echo "Started compressing $output_name."
      upx --best --lzma ./build/$output_name
      if [ $? -ne 0 ]; then
        echo 'An error has occurred! Aborting the script execution...'
        exit 1
      fi
      echo "Finished compressing $output_name."
    fi
  fi
done
