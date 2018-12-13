
rm -rf ./out
rm -rf ./ipfs-monitor-out*

VERSION='0.1.0b'

mkdir out
mkdir ./out/amd64
FILENAME=iphash-windows-amd64-v$VERSION
CURRDIR=./out/amd64/$FILENAME
mkdir $CURRDIR
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build  -o  $CURRDIR/ipfs-monitor.exe
cp -f ./temp/script/windows/* $CURRDIR/
cp -f ./temp/key/swarm.key $CURRDIR/
cp -f ./temp/ipfs/ipfs0.4.18.exe $CURRDIR/ipfs.exe
cp -f ./temp/ipfs/fs-repo-migrations.exe $CURRDIR/fs-repo-migrations.exe
mv $CURRDIR ./$FILENAME
tar -zcvf $CURRDIR.tar.gz ./$FILENAME
rm -r $FILENAME/
shasum $CURRDIR.tar.gz >> $CURRDIR.sha1.txt

FILENAME=iphash-linux-amd64-v$VERSION
CURRDIR=./out/amd64/$FILENAME
mkdir $CURRDIR
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build  -o  $CURRDIR/ipfs-monitor
cp -f ./temp/script/unix/* $CURRDIR/
cp -f ./temp/key/swarm.key $CURRDIR/
cp -f ./temp/ipfs/ipfs-amd $CURRDIR/ipfs
cp -f ./temp/ipfs/fs-repo-migrations-linux-amd64 $CURRDIR/fs-repo-migrations
mv $CURRDIR ./$FILENAME
tar -zcvf $CURRDIR.tar.gz ./$FILENAME
rm -r $FILENAME/
shasum $CURRDIR.tar.gz >> $CURRDIR.sha1.txt


FILENAME=iphash-darwin-amd64-v$VERSION
CURRDIR=./out/amd64/$FILENAME
mkdir $CURRDIR
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build  -o  $CURRDIR/ipfs-monitor
cp -f ./temp/script/unix/* $CURRDIR/
cp -f ./temp/key/swarm.key $CURRDIR/
cp -f ./temp/ipfs/ipfs-darwin $CURRDIR/ipfs
mv $CURRDIR ./$FILENAME
tar -zcvf $CURRDIR.tar.gz ./$FILENAME
rm -r $FILENAME/
shasum $CURRDIR.tar.gz >> $CURRDIR.sha1.txt

mkdir ./out/arm64

FILENAME=iphash-linux-arm64-v$VERSION
CURRDIR=./out/arm64/$FILENAME
mkdir $CURRDIR
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build  -o  $CURRDIR/ipfs-monitor
cp -f ./temp/script/unix/* $CURRDIR/
cp -f ./temp/key/swarm.key $CURRDIR/
cp -f ./temp/ipfs/ipfs-arm $CURRDIR/ipfs
cp -f ./temp/ipfs/fs-repo-migrations-linux-arm64 $CURRDIR/fs-repo-migrations
mv $CURRDIR ./$FILENAME
tar -zcvf $CURRDIR.tar.gz ./$FILENAME
rm -r $FILENAME/
shasum $CURRDIR.tar.gz >> $CURRDIR.sha1.txt

zip -r ipfs-monitor-out-$(date +"%Y%m%d%H%M%s").zip ./out
