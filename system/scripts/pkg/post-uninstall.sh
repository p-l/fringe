#!/bin/bash

function disable_systemd {
    systemctl disable fringe
    rm -f /lib/systemd/system/fringe.service
}

function disable_update_rcd {
    update-rc.d -f fringe remove
    rm -f /etc/init.d/fringe
}

function disable_chkconfig {
    chkconfig --del fringe
    rm -f /etc/init.d/fringe
}

if [[ -f /etc/redhat-release ]]; then
    # RHEL-variant logic
    if [[ "$1" = "0" ]]; then
        # fringe is no longer installed, remove from init system
        rm -f /etc/default/fringe

        if command -v systemctl &>/dev/null; then
            disable_systemd
        else
            # Assuming sysv
            disable_chkconfig
        fi
    fi
elif [[ -f /etc/lsb-release ]]; then
    # Debian/Ubuntu logic
    if [[ "$1" != "upgrade" ]]; then
        # Remove/purge
        rm -f /etc/default/fringe

        if command -v systemctl &>/dev/null; then
            disable_systemd
        else
            # Assuming sysv
            disable_update_rcd
        fi
    fi
elif [[ -f /etc/os-release ]]; then
    source /etc/os-release
    if [[ "$ID" = "amzn" ]] && [[ "$1" = "0" ]]; then
        # fringe is no longer installed, remove from init system
        rm -f /etc/default/fringe

        if [[ "$NAME" = "Amazon Linux" ]]; then
            # Amazon Linux 2+ logic
            disable_systemd
        elif [[ "$NAME" = "Amazon Linux AMI" ]]; then
            # Amazon Linux logic
            disable_chkconfig
        fi
    fi
fi