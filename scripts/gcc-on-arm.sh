#!/bin/sh
# Install GCC on ARM architecture, used by Docker.

case $1 in
	"-h" | "--help" ) echo "usage: $(basename "$0")"; exit;;
esac

if [ $# -gt 0 ]; then
	echo "error: wrong number of arguments" 1>&2
	exit 1
fi

case $(uname -m) in
	arm* )
		apk add gcc
		;;
	* )
		echo "not arm"
		;;
esac
