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

OSES=(linux darwin windows freebsd)
ARCHS=(amd64 386)
for os in ${OSES[@]}; do
	for arch in ${ARCHS[@]}; do
		suffix=""
		if [ "$os" == "windows" ]
		then
			suffix=".exe"
		fi
		pushd client
		env CGO_ENABLED=0 GOOS=$os GOARCH=$arch go build -ldflags "$LDFLAGS" -gcflags "$GCFLAGS" -o ../$OUT/client_${os}_${arch}${suffix} .
		popd

		pushd server
		env CGO_ENABLED=0 GOOS=$os GOARCH=$arch go build -ldflags "$LDFLAGS" -gcflags "$GCFLAGS" -o ../$OUT/server_${os}_${arch}${suffix} .
		popd

		pushd $OUT
		if $UPX; then upx -9 client_${os}_${arch}${suffix} server_${os}_${arch}${suffix};fi
		tar -zcf kcptun-${os}-${arch}-$VERSION.tar.gz client_${os}_${arch}${suffix} server_${os}_${arch}${suffix}
		$sum kcptun-${os}-${arch}-$VERSION.tar.gz
		popd
	done
done

# ARM
ARMS=(5 6 7)
for v in ${ARMS[@]}; do
	pushd client
	env CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=$v go build -ldflags "$LDFLAGS" -gcflags "$GCFLAGS" -o ../$OUT/client_linux_arm$v  .
	popd

	pushd server
	env CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=$v go build -ldflags "$LDFLAGS" -gcflags "$GCFLAGS" -o ../$OUT/server_linux_arm$v  .
	popd
done
pushd $OUT
if $UPX; then upx -9 client_linux_arm* server_linux_arm*;fi
tar -zcf kcptun-linux-arm-$VERSION.tar.gz client_linux_arm* server_linux_arm*
$sum kcptun-linux-arm-$VERSION.tar.gz
popd

# ARMv8/ARM64/aarch64
pushd client
env CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags "$LDFLAGS" -gcflags "$GCFLAGS" -o ../$OUT/client_linux_arm64  .
popd

pushd server
env CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags "$LDFLAGS" -gcflags "$GCFLAGS" -o ../$OUT/server_linux_arm64  .
popd

pushd $OUT
if $UPX; then upx -9 client_linux_arm64 server_linux_arm64;fi
tar -zcf kcptun-linux-arm64-$VERSION.tar.gz client_linux_arm64 server_linux_arm64
$sum kcptun-linux-arm64-$VERSION.tar.gz
popd



#MIPS32LE
MIPS=(mipsle mips)
for v in ${MIPS[@]}; do
	pushd client
	env CGO_ENABLED=0 GOOS=linux GOARCH=$v GOMIPS=softfloat go build -ldflags "$LDFLAGS" -gcflags "$GCFLAGS" -o ../$OUT/client_linux_$v .
	popd

	pushd server
	env CGO_ENABLED=0 GOOS=linux GOARCH=$v GOMIPS=softfloat go build -ldflags "$LDFLAGS" -gcflags "$GCFLAGS" -o ../$OUT/server_linux_$v .
	popd

	pushd $OUT
	if $UPX; then upx -9 client_linux_$v server_linux_$v;fi
	tar -zcf kcptun-linux-$v-$VERSION.tar.gz client_linux_mips* client_linux_mips*
	$sum kcptun-linux-$v-$VERSION.tar.gz
	popd
done

