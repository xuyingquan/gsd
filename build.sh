#!/bin/bash
export GOPATH=`pwd`:`pwd`/../GolangDeps
export WORKDIR=`pwd`
echo "get dependency"
#go get github.com/mgutz/logxi/v1
#go get github.com/miekg/dns
#go get gopkg.in/redis.v3
#go get stathat.com/c/consistent
echo "build gsd5"
rm -f bin/gsd.bin
rm -f bin/gsd
go build -o bin/gsd.bin src/main.go
if [[ $? != 0 ]]; then
    echo "\033[31mbuild gsd failed\033[0m" >&2
    exit $?
fi
cp src/gsd bin/gsd
echo "package"
rm -rf gsd5
mkdir -p $WORKDIR/gsd5
cp -R bin $WORKDIR/gsd5/
cp -R etc $WORKDIR/gsd5/
mkdir -p $WORKDIR/gsd5/logs
rm -f gsd5.tar.gz
cd gsd5
tar czvf gsd5.tar.gz *
if [[ $? != 0 ]]; then
    echo "\033[31mpackage gsd failed\033[0m" >&2
	exit $?
fi
mv gsd5.tar.gz ../
cd ..
rm -rf gsd5
echo "build ipfrom"
rm -f bin/ipfrom
go build -o bin/ipfrom src/ipfrom.go
if [[ $? != 0 ]]; then
    echo "\033[31mbuild ipfrom failed\033[0m" >&2
    exit $?
fi
echo "build dnsbench"
go build -o bin/dnsbench src/dnsbench.go
if [[ $? != 0 ]]; then
    echo "\033[31mbuild dnsbench failed\033[0m" >&2
    exit $?
fi
echo "OK,exec file is bin/gsd, package file is build/gsd.tar.gz"
