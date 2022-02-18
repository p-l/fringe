#!/bin/bash

set -e

if [ -z "${1}" ] || [ -z "${2}" ]; then
    echo "usage ${0} <repo root dir> <arch>"
    exit 1
fi

REPO_ROOT_DIR=${1}
REPO_ARCH=${2}

mkdir -p ${REPO_ROOT_DIR}/dists/stable/main/binary-${REPO_ARCH}/
cd ${REPO_ROOT_DIR}
dpkg-scanpackages --arch ${REPO_ARCH} pool/ > dists/stable/main/binary-${REPO_ARCH}//Packages
cat dists/stable/main/binary-${REPO_ARCH}/Packages | gzip -9 > dists/stable/main/binary-${REPO_ARCH}/Packages.gz
