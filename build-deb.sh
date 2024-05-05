#!/bin/bash

msg-run() {
	echo "> $*"
	"$@"
}

# Install build dependnecies.
#msg-run sudo apt-get build-dep .

msg-run dpkg-buildpackage -b

