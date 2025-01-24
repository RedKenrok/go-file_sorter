GOOS=windows GOARCH=amd64 \
go build -ldflags="-s -w" -o="./build/windows/image_sorter.exe" ./main.go

upx --best --ultra-brute ./build/windows/image_sorter.exe