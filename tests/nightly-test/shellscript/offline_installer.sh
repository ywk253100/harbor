#!/bin/sh

set -e

installer_dir="/harbor"
if [ -d $installer_dir ]; then    
    rm -rf $installer_dir/*
else
    mkdir -p $installer_dir
fi

CUR=$PWD

generate_ca() {
    ./$PWD/tests/nightly-test/shellscript/ca_generator.sh $1
    # ./ca_generator.sh $1
}

get_installer() {
    local target_url=""
    if [ $1='latest' ]; then
        target_url="https://storage.googleapis.com/harbor-builds/harbor-offline-installer-latest.tgz"
        echo "URL is not specified. Default url is latest."
    else
        target_url=$1
    fi
    curl -O $target_url
    mv ./harbor-offline-installer-latest.tgz $installer_dir
    cd $installer_dir
    tar -zvxf ./harbor-offline-installer-latest.tgz
    cd $CUR
}

set_harbor_cfg() {
    sed "s/reg.mydomain.com/$IP/" -i $installer_dir/harbor/harbor.cfg
    python ./$PWD/tests/nightly-test/configuration/edit-cfg.py --config $installer_dir/harbor/harbor.cfg --in-json ./$PWD/tests/nightly-test/configuration/$1.json
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