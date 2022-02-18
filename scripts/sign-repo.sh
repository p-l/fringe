if [ -z "${SIGN_IDENTITY}" ]; then
    echo "SIGN_IDENTITY not defined. Skipping signature"
    exit 0
fi

if [ -z "${1}" ]; then
    echo "Usage: ${0} <repo root dir>"
    exit 1
fi
REPO_ROOT_DIR=${1}

# Export repository key
gpg --armor --export ${SIGN_IDENTITY} > ${REPO_ROOT_DIR}/repo.key

# Sign the Release
cat ${REPO_ROOT_DIR}/dists/stable/Release | gpg --default-key ${SIGN_IDENTITY} -abs --clearsign > ${REPO_ROOT_DIR}/dists/stable/InRelease
