#!/bin/sh
set -e

if [ -z "${1}" ]; then
    echo "usage ${0} <repo root dir>"
    exit 1
fi
REPO_ROOT_DIR=${1}

cd ${REPO_ROOT_DIR}/dists/stable

do_hash() {
    HASH_NAME=$1
    HASH_CMD=$2
    echo "${HASH_NAME}:"
    for f in $(find -type f); do
        f=$(echo $f | cut -c3-) # remove ./ prefix
        if [ "$f" = "Release" ]; then
            continue
        fi
        echo " $(${HASH_CMD} ${f}  | cut -d" " -f1) $(wc -c $f)"
    done
}

cat << EOF
Origin: Fringe Repository 
Label: fringe
Suite: stable
Codename: stable
Version: 1.0
Architectures: amd64 arm64 arm 386
Components: main
Description: Repository for Fringe packages
Date: $(date -Ru)
EOF
do_hash "MD5Sum" "md5sum"
do_hash "SHA1" "sha1sum"
do_hash "SHA256" "sha256sum"

