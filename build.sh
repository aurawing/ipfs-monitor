
rm -rf ./out
mkdir out
mkdir ./out/amd64
mkdir ./out/amd64/windows
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build  -o  ./out/amd64/windows/ipfs-monitor.exe
cp -f ./temp/script/windows/* ./out/amd64/windows/
cp -f ./temp/key/swarm.key ./out/amd64/windows/
cp -f ./temp/ipfs/ipfs.exe ./out/amd64/windows/ipfs.exe

mkdir ./out/amd64/linux
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o  ./out/amd64/linux/ipfs-monitor
cp -f ./temp/script/unix/* ./out/amd64/linux/
cp -f ./temp/key/swarm.key ./out/amd64/linux/
cp -f ./temp/ipfs/ipfs-amd ./out/amd64/linux/ipfs


mkdir ./out/amd64/darwin
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o  ./out/amd64/darwin/ipfs-monitor

mkdir ./out/arm64

mkdir ./out/arm64/linux
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o  ./out/arm64/linux/ipfs-monitor
cp -f ./temp/script/unix/* ./out/arm64/linux/
cp -f ./temp/key/swarm.key ./out/arm64/linux/
cp -f ./temp/ipfs/ipfs-arm ./out/arm64/linux/ipfs

zip -r ipfs-monitor-out-$(date +"%Y%m%d%H%M%s").zip ./out
