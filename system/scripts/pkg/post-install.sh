#!/bin/bash

BIN_DIR=/usr/bin
DATA_DIR=/var/lib/fringe
LOG_DIR=/var/log/fringe
SCRIPT_DIR=/usr/lib/fringe/scripts
LOGROTATE_DIR=/etc/logrotate.d
FRINGE_CONFIG_PATH=/etc/fringe/config.toml

function install_init {
    cp -f $SCRIPT_DIR/init.sh /etc/init.d/fringe
    chmod +x /etc/init.d/fringe
}

function install_systemd {
    cp -f $SCRIPT_DIR/fringe.service /lib/systemd/system/fringe.service
    systemctl enable fringe
}

function install_update_rcd {
    update-rc.d fringe defaults
}

function install_chkconfig {
    chkconfig --add fringe
}

function should_upgrade {
    return 0
}

function upgrade_notice {
cat << EOF
EOF
}

function init_config {
    mkdir -p "$(dirname ${FRINGE_CONFIG_PATH})"

    local config_path=${FRINGE_CONFIG_PATH}
    if [[ -s ${config_path} ]]; then
        config_path=${FRINGE_CONFIG_PATH}.defaults
        echo "Config file ${FRINGE_CONFIG_PATH} already exists, writing defaults to ${config_path}"
    fi
}

# Add defaults file, if it doesn't exist
if [[ ! -s /etc/default/fringe ]]; then
cat << EOF > /etc/default/fringe
FRINGE_CONFIG_PATH=${FRINGE_CONFIG_PATH}
EOF
fi

# Remove legacy symlink, if it exists
if [[ -L /etc/init.d/fringe ]]; then
    rm -f /etc/init.d/fringe
fi

# Distribution-specific logic
if [[ -f /etc/redhat-release ]]; then
    # RHEL-variant logic
    if command -v systemctl &>/dev/null; then
        install_systemd
    else
        # Assuming sysv
        install_init
        install_chkconfig
    fi
elif [[ -f /etc/debian_version ]]; then
    # Ownership for RH-based platforms is set in build.py via the `rmp-attr` option.
    # We perform ownership change only for Debian-based systems.
    # Moving these lines out of this if statement would make `rmp -V` fail after installation.
    chown -R -L fringe:fringe $LOG_DIR
    chown -R -L fringe:fringe $DATA_DIR
    chmod 755 $LOG_DIR
    chmod 755 $DATA_DIR

    # Debian/Ubuntu logic
    if command -v systemctl &>/dev/null; then
        install_systemd
    else
        # Assuming sysv
        install_init
        install_update_rcd
    fi
elif [[ -f /etc/os-release ]]; then
    source /etc/os-release
    if [[ "$NAME" = "Amazon Linux" ]]; then
        # Amazon Linux 2+ logic
        install_systemd
    elif [[ "$NAME" = "Amazon Linux AMI" ]]; then
        # Amazon Linux logic
        install_init
        install_chkconfig
    fi
fi

# Check upgrade status
if should_upgrade; then
    upgrade_notice
else
    init_config
fi