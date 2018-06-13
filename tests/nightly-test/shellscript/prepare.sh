#!/bin/sh

set -e

function create_user {
    curl -u admin:Harbor12345 -k -X POST --header 'Content-Type: application/json' --header 'Accept: text/plain' -d '{
        "username": "'$1'",
        "email": "'$1@vmware.com'",
        "password": "Test1@34",
        "realname": "user for testing",
        "comment": "user comment for testing"
        }' 'https://'$2'/api/users'
}

create_user user001 $1
create_user user002 $1
create_user user003 $1
create_user user004 $1
create_user user005 $1
create_user user006 $1
create_user user007 $1
create_user user008 $1
create_user user009 $1
create_user user010 $1
create_user user011 $1
create_user user012 $1
create_user user013 $1
create_user user014 $1
create_user user015 $1
create_user user016 $1
create_user user017 $1
create_user user018 $1
create_user user019 $1
create_user user020 $1
create_user user021 $1
create_user user022 $1
create_user user023 $1
create_user user024 $1
create_user user025 $1
create_user user026 $1
