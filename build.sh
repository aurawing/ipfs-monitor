
rm -rf ./out
mkdir out
mkdir ./out/amd64
mkdir ./out/amd64/windows
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build  -o  ./out/amd64/windows/ipfs-monitor.exe

mkdir ./out/amd64/linux
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o  ./out/amd64/linux/ipfs-monitor

mkdir ./out/amd64/darwin
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o  ./out/amd64/darwin/ipfs-monitor

mkdir ./out/arm64

mkdir ./out/arm64/linux
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o  ./out/arm64/linux/ipfs-monitor