#!/bin/sh

set -e

umask 077

data_dir="/data"
ip_address=$1

# Copy CA cert to be downloaded from UI
ca_download_dir="${data_dir}/ca_download"
rm -rf "${ca_download_dir}"
mkdir -p "${ca_download_dir}"

cert="${data_dir}/cert/server.crt"
key="${data_dir}/cert/server.key"

generateCerts() {
  # Create CA certificate
  openssl req \
      -newkey rsa:4096 -nodes -sha256 -keyout harbor_ca.key \
      -x509 -days 365 -out harbor_ca.crt -subj '/C=CN/ST=PEK/L=BeiJing/O=VMware/CN=HarborCA'
  
  # Generate a Certificate Signing Request
  openssl req \
      -newkey rsa:4096 -nodes -sha256 -keyout $ip_address.key \
      -out $ip_address.csr -subj "/C=CN/ST=PEK/L=BeiJing/O=VMware/CN=HarborManager"
  
  # Generate the certificate of local registry host
  echo subjectAltName = IP:$ip_address > extfile.cnf
  openssl x509 -req -days 365 -in $ip_address.csr -CA harbor_ca.crt \
    -CAkey harbor_ca.key -CAcreateserial -extfile extfile.cnf -out $ip_address.crt      


  # Copy to harbor default location
  mkdir -p /data/cert
  cp $ip_address.crt $cert
  cp $ip_address.key $key
  cp harbor_ca.crt ${ca_download_dir}/ca.crt
  chown --recursive 10000:10000 ${ca_download_dir}
}

generateCerts