#!/bin/bash

# Inspired from:
# https://github.com/influxdata/fringe/blob/c80c3636dd7bba6c3d0f3d6a1c321862da8f201a/releng/packages/fs/usr/local/bin/fringe_packages.bash
set -e

ARTIFACTS_DIR="artifacts"
OUT_DIR="$ARTIFACTS_DIR"/out
PKG_DIR="$ARTIFACTS_DIR"/pkg
BIN_DIR="bin"
BIN_FILE="fringe-server"
version=$(git rev-parse --short HEAD)
VERSION_STRING="$(cat VERSION)-${version}"

DEB_PACKAGE_DESCRIPTION="Radius bridge for Google OAuth"
DEB_PACKAGE_NAME="fringe"
DEB_PACKAGE_URL="https://github.com/p-l/fringe"

# check all the required environment variables are supplied
[ -z "$GOOS" ] && echo "Need to set GOOS to target linux" && exit
[ -z "$GOARCH" ] && echo "Need to set GOARCH" && exit

# Build the binary
make GOOS="$GOOS" GOARCH="$GOARCH" build

# Create package directory
mkdir -p $ARTIFACTS_DIR \
         "$OUT_DIR" \
         "$PKG_DIR" \
         "$PKG_DIR"/usr/bin \
         "$PKG_DIR"/usr/lib/fringe \
         "$PKG_DIR"/usr/lib/fringe/scripts \
         "$PKG_DIR"/var/log/fringe \
         "$PKG_DIR"/var/lib/fringe \
         "$PKG_DIR"/etc/fringe \
         "$PKG_DIR"/etc/logrotate.d
chmod -R 0755 "$PKG_DIR"

# Copy binary
cp $BIN_DIR/"$BIN_FILE-$GOOS-$GOARCH" "$PKG_DIR"/usr/bin/fringe

# Copy service scripts
cp system/scripts/service/init.sh "$PKG_DIR"/usr/lib/fringe/scripts/init.sh
cp system/scripts/service/fringe.service "$PKG_DIR"/usr/lib/fringe/scripts/fringe.service
chmod 0644 "$PKG_DIR"/usr/lib/fringe/scripts/fringe.service
cp system/scripts/service/fringe-systemd.sh "$PKG_DIR"/usr/lib/fringe/scripts/fringe-systemd.sh

# Copy logrotate.d script
cp system/scripts/logrotate.d/logrotate "$PKG_DIR"/etc/logrotate.d/fringe
chmod 0644 "$PKG_DIR"/etc/logrotate.d/fringe

# Copy sample config
cp config/sample-config.toml "$PKG_DIR"/etc/fringe/config.toml

# NOTE:
# fpm depends on a number of other binary not tested in the script.
# for debian package gnu-tar is required if you're building on macOS
# brew install gnu-tar then add it at the begin of the path with
#   export PATH=$(brew --prefix gnu-tar)/libexec/gnubin:${PATH}
PACKAGE_FILENAME="${DEB_PACKAGE_NAME}_${VERSION_STRING}_${GOARCH}".deb
bundle exec fpm --verbose \
  --name "$DEB_PACKAGE_NAME" \
  --description "$DEB_PACKAGE_DESCRIPTION" \
  --url "$DEB_PACKAGE_URL" \
  --version "$VERSION_STRING" \
  --license LICENSE.md \
  --architecture "$GOARCH" \
  --iteration 1 \
  --output-type deb \
  --input-type dir \
  --before-install system/scripts/pkg/pre-install.sh \
  --after-install system/scripts/pkg/post-install.sh \
  --after-remove system/scripts/pkg/post-uninstall.sh \
  --chdir "$PKG_DIR" \
  -p "$OUT_DIR"/"$PACKAGE_FILENAME"

mv "$OUT_DIR"/"$PACKAGE_FILENAME" "$ARTIFACTS_DIR/"
rm -rf "$OUT_DIR"
rm -rf "$PKG_DIR"

md5sum "$ARTIFACTS_DIR"/"$PACKAGE_FILENAME" > "$ARTIFACTS_DIR"/"$PACKAGE_FILENAME".md5
sha256sum "$ARTIFACTS_DIR"/"$PACKAGE_FILENAME" > "$ARTIFACTS_DIR"/"$PACKAGE_FILENAME".sha256
