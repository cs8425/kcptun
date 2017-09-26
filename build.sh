#!/bin/bash

OUT="./bin/"

VERSION=`date -u +%Y%m%d%H%M`
LDFLAGS="-X main.VERSION=$VERSION -s -w"
GCFLAGS=""

ARCHS=(x64 x86 arm7 win64 arm8 mipsle)
for v in ${ARCHS[@]}; do
	pushd client/
	go-$v build -o client-kcp.$v -ldflags "$LDFLAGS" -gcflags "$GCFLAGS" .
	popd
	mv client/client-kcp.$v $OUT


	pushd server/
	go-$v build -o server-kcp.$v -ldflags "$LDFLAGS" -gcflags "$GCFLAGS" .
	popd
	mv server/server-kcp.$v $OUT

done


