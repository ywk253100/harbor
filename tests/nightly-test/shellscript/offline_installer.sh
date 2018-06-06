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
    ./tests/nightly-test/shellscript/ca_generator.sh $1
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
    sed "s/reg.mydomain.com/$1/" -i $installer_dir/harbor/harbor.cfg
    python ./tests/nightly-test/configuration/edit-cfg.py --config $installer_dir/harbor/harbor.cfg --in-json ./tests/nightly-test/configuration/$2.json
}

# have notary and clair installed.
install() {
    cd $installer_dir/harbor
    ./install.sh --with-notary --with-clair 
}

clean_up() {
    curl --insecure -s -L -H "Accept: application/json" https://$1/ | grep "Harbor"  > /dev/null
    if [ $? -eq 0 ]; then 
        echo "Harbor is not running on $1"
        cd $installer_dir/harbor
        docker-compose -f docker-compose.yml -f docker-compose.notary.yml -f docker-compose.clair.yml down -v
    fi
    
    # Clean data...
    cd /data
    rm -rf ./*
    cd /var/log/harbor
    rm -rf ./*
    cd $CUR
}

main() {
    local auth_type=$1
    local ip_address=$2
    local url=$3

    clean_up $2
    get_installer $url
    generate_ca $ip_address
    set_harbor_cfg $ip_address $auth_type
    install
}

main "$@"
