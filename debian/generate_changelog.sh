#!/usr/bin/bash

export DEBEMAIL="info@specterops.io"
export PKG_VERSION=$(git describe --tags --match "v*" | sed 's/^v//g')

dch -i "$PKG_VERSION.0" \
	"Automatic version bump"
