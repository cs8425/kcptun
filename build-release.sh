#!/bin/bash
export GO111MODULE=on
sum="sha1sum"

if ! hash sha1sum 2>/dev/null; then
	if ! hash shasum 2>/dev/null; then
		echo "I can't see 'sha1sum' or 'shasum'"
		echo "Please install one of them!"
		exit
	fi
	sum="shasum"
fi

UPX=false
if hash upx 2>/dev/null; then
	UPX=true
fi

VERSION=`git describe --tags`
LDFLAGS="-X main.VERSION=$VERSION -s -w"
GCFLAGS=""
OUT="build"

startgo() {
NAME=$1
shift
case "$NAME" in
'freebsd-amd64')
	rungo freebsd amd64 7 "$@"
;;
'freebsd-386')
	rungo freebsd 386 7 "$@"
;;

'darwin-amd64')
	rungo darwin amd64 7 "$@"
;;
'darwin-386')
	rungo darwin 386 7 "$@"
;;

'amd64')
	rungo linux amd64 7 "$@"
;;
'386')
	rungo linux 386 7 "$@"
;;
'arm5')
	rungo linux arm 5 "$@"
;;
'arm6')
	rungo linux arm 6 "$@"
;;
'arm7')
	rungo linux arm 7 "$@"
;;
'arm8' | 'arm64' | 'aarch64')
	rungo linux arm64 7 "$@"
;;
'mipsle')
	rungo linux mipsle 7 "$@"
;;
'mips')
	rungo linux mips 7 "$@"
;;

'windows-386')
	rungo windows 386 7 "$@"
;;
'windows-amd64')
	rungo windows amd64 7 "$@"

esac
}

rungo() {
	OS=$1
	shift
	ARCH=$1
	shift
	ARM=$1
	shift
	#echo "[$OS, $ARCH, $ARM]": "$@"
	CGO_ENABLED=0 GOOS=$OS GOARCH=$ARCH GOARM=$ARM go "$@"
}


ARCHS=(amd64 386 darwin-amd64 darwin-386 windows-386 windows-amd64 freebsd-amd64 freebsd-386 arm8 arm7 arm6 arm5 mipsle mips)
for v in ${ARCHS[@]}; do
	suffix=""

	arch=`echo $v|cut -d'-' -f2`
	os=`echo $v|cut -d'-' -f1`
	if [ "$os" == "$arch" ]
	then
		os="linux"
	fi
	if [ "$os" == "windows" ]
	then
		suffix=".exe"
	fi

	pushd client
	startgo $v build -ldflags "$LDFLAGS" -gcflags "$GCFLAGS" -o ../$OUT/client_${os}_${arch}${suffix} .
	popd

	pushd server
	startgo $v build -ldflags "$LDFLAGS" -gcflags "$GCFLAGS" -o ../$OUT/server_${os}_${arch}${suffix} .
	popd

	pushd $OUT
	if $UPX; then upx -9 client_${os}_${arch}${suffix} server_${os}_${arch}${suffix};fi
	tar -zcf kcptun-${os}-${arch}-$VERSION.tar.gz client_${os}_${arch}${suffix} server_${os}_${arch}${suffix}
	$sum kcptun-${os}-${arch}-$VERSION.tar.gz
	popd
done

