
rm -rf ./out
rm -rf ./ipfs-monitor-out*

VERSION='0.1.0b'

mkdir out
mkdir ./out/amd64
FILENAME=./out/amd64/iphash-windows-amd64-v$VERSION
mkdir $FILENAME
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build  -o  $FILENAME/ipfs-monitor.exe
cp -f ./temp/script/windows/* $FILENAME/
cp -f ./temp/key/swarm.key $FILENAME/
cp -f ./temp/ipfs/ipfs.exe $FILENAME/ipfs.exe
tar -zcvf ./out/amd64/iphash-windows-amd64-v$VERSION.tar.gz $FILENAME
rm -r $FILENAME/
shasum ./out/amd64/iphash-windows-amd64-v$VERSION.tar.gz >> $FILENAME.sha1.txt

FILENAME=./out/amd64/iphash-linux-amd64-v$VERSION
mkdir $FILENAME
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build  -o  $FILENAME/ipfs-monitor
cp -f ./temp/script/unix/* $FILENAME/
cp -f ./temp/key/swarm.key $FILENAME/
cp -f ./temp/ipfs/ipfs-amd $FILENAME/ipfs
tar -zcvf ./out/amd64/iphash-linux-amd64-v$VERSION.tar.gz $FILENAME
rm -r $FILENAME/
shasum ./out/amd64/iphash-linux-amd64-v$VERSION.tar.gz >> $FILENAME.sha1.txt


FILENAME=./out/amd64/iphash-darwin-amd64-v$VERSION
mkdir $FILENAME
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build  -o  $FILENAME/ipfs-monitor
cp -f ./temp/script/unix/* $FILENAME/
cp -f ./temp/key/swarm.key $FILENAME/
cp -f ./temp/ipfs/ipfs-amd $FILENAME/ipfs
tar -zcvf ./out/amd64/iphash-darwin-amd64-v$VERSION.tar.gz $FILENAME
rm -r $FILENAME/
shasum ./out/amd64/iphash-darwin-amd64-v$VERSION.tar.gz >> $FILENAME.sha1.txt

mkdir ./out/arm64

FILENAME=./out/arm64/iphash-linux-arm64-v$VERSION
mkdir $FILENAME
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build  -o  $FILENAME/ipfs-monitor
cp -f ./temp/script/unix/* $FILENAME/
cp -f ./temp/key/swarm.key $FILENAME/
cp -f ./temp/ipfs/ipfs-amd $FILENAME/ipfs
tar -zcvf ./out/arm64/iphash-linux-arm64-v$VERSION.tar.gz $FILENAME
rm -r $FILENAME/
shasum ./out/arm64/iphash-linux-arm64-v$VERSION.tar.gz >> $FILENAME.sha1.txt

zip -r ipfs-monitor-out-$(date +"%Y%m%d%H%M%s").zip ./out
