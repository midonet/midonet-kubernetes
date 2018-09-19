#! /bin/sh

set -e

CMD=$1

VARIANTS="amd64-linux arm64v8-linux"

run() {
	if [ "$CMD" = "check" ]; then
		m4 -I m4/inc m4/$1.m4 > $1.tmp
		diff -upd $1 $1.tmp
		rm $1.tmp
	else
		m4 -I m4/inc m4/$1.m4 > $1
	fi
}

for VARIANT in $VARIANTS; do
	run Dockerfile-${VARIANT}
	run Dockerfile-node-${VARIANT}
done
run Dockerfile-test
