#!/bin/sh

set -e

installer_dir="/harbor"
if [ -d $installer_dir ]; then    
    rm -rf ./$installer_dir/*
else
    mkdir -p $installer_dir
fi

generate_ca() {
    ./ca_generator.sh
}

get_installer() {
    local target_url=""
    if [ -z $1 ]; then
        target_url="https://storage.googleapis.com/harbor-builds/harbor-offline-installer-latest.tgz"
        echo "URL is not specified. Default url is latest."
    else
        target_url=$1
    fi
    curl -O $target_url
    mv ./harbor-offline-installer-latest.tgz $installer_dir
    tar -zvxf $installer_dir/harbor-offline-installer-latest.tgz
}

set_harbor_cfg() {
    python ../configuration/edit-cfg.py --config $installer_dir/harbor/harbor.cfg --in-json $2.json
}

# have notary and clair installed.
install() {
    $installer_dir/harbor/install.sh --with-notary --with-clair 
}

main() {
    local auth_type=$1
    local ip_address=$2
    local url=$3

    get_installer $url
    generate_ca $ip_address
    set_harbor_cfg $auth_type
    install
}

main "$@"