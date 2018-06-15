#!/bin/sh

set -e

installer_dir="/harbor"
CUR=$PWD

generate_ca() {
    ./tests/nightly-test/shellscript/ca_generator.sh $1
}

get_installer() {
    local target_url=""
    if [ "$1" = 'latest' ]; then
        target_url="https://storage.googleapis.com/harbor-builds/harbor-offline-installer-latest.tgz"
        echo "URL is not specified. Default url is latest."
    else
        target_url=$1
    fi
    curl -O $target_url
    harbor_build_bundle=$(basename ./harbor-offline-installer-*.tgz)
    mv ./$harbor_build_bundle $installer_dir
    cd $installer_dir
    tar -zvxf ./$harbor_build_bundle
    cd $CUR
}

set_harbor_cfg() {
    sed "s/reg.mydomain.com/$1/" -i ./tests/nightly-test/configuration/$2.json
    sed "s/#ldap_searchdn/ldap_searchdn/" -i $installer_dir/harbor/harbor.cfg
    sed "s/#ldap_search_pwd/ldap_search_pwd/" -i $installer_dir/harbor/harbor.cfg
    sed "s/#ldap_filter/ldap_filter/" -i $installer_dir/harbor/harbor.cfg
    python ./tests/nightly-test/configuration/edit-cfg.py --config $installer_dir/harbor/harbor.cfg --in-json ./tests/nightly-test/configuration/$2.json
}

# have notary and clair installed.
install() {
    cd $installer_dir/harbor
    ./install.sh --with-notary --with-clair 
}

init() {
    if [ -d $installer_dir/harbor ]; then
        cd $installer_dir/harbor
        if [ -n "$(docker-compose -f docker-compose.yml -f docker-compose.notary.yml -f docker-compose.clair.yml ps -q)" ]; then
            docker-compose -f docker-compose.yml -f docker-compose.notary.yml -f docker-compose.clair.yml down -v
        fi
        rm -rf $installer_dir/harbor/*
    else
        mkdir -p $installer_dir
    fi
    
    # Clean data...
    cd /data
    rm -rf ./*
    cd /var/log/harbor
    rm -rf ./*
    cd /harbor/ca
    rm -rf ./*
    cd $CUR
}

main() {
    local auth_type=$1
    local ip_address=$2
    local url=$3

    init
    get_installer $url
    generate_ca $ip_address
    set_harbor_cfg $ip_address $auth_type
    install
}

main "$@"
