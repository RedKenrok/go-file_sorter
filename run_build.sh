#!/usr/bin/env bash

package_name="file_sorter"

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

for platform in "${platforms[@]}"
do
  platform_split=(${platform//\// })
  GOOS=${platform_split[0]}
  GOARCH=${platform_split[1]}
  output_name=$package_name'-'$GOOS'-'$GOARCH
  if [ $GOOS = "windows" ]; then
    output_name+='.exe'
  fi

  echo "Started building $output_name."
  env CGO_ENABLED=0 GOOS=$GOOS GOARCH=$GOARCH go build -ldflags="-s -w" -o="./build/$output_name" ./main.go
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
