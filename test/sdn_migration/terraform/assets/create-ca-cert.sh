#!/bin/bash

export GIT_ROOT=$(git rev-parse --show-toplevel)/$(git rev-parse --show-prefix)

openssl genrsa -out ${GIT_ROOT}/ca.key 4096
openssl req -new -newkey rsa:4096 -nodes -keyout ${GIT_ROOT}/ca.key -out ${GIT_ROOT}/ca.csr -subj "/C=XX/ST=XX/L=squid/O=squid/CN=squid"

cat <<EOF>${GIT_ROOT}/root-ca.cnf
[root_ca]
basicConstraints = critical,CA:TRUE,pathlen:1
keyUsage = critical, nonRepudiation, cRLSign, keyCertSign
subjectKeyIdentifier=hash
EOF

openssl x509 -req -days 3650 -in ${GIT_ROOT}/ca.csr -signkey ${GIT_ROOT}/ca.key -out ${GIT_ROOT}/ca.crt -extfile ${GIT_ROOT}/root-ca.cnf -extensions root_ca

cat ${GIT_ROOT}/ca.key ${GIT_ROOT}/ca.crt > ${GIT_ROOT}/ca-bundle.crt

openssl x509 -in ${GIT_ROOT}/ca-bundle.crt -outform PEM -out ${GIT_ROOT}/ca.pem

rm -f ${GIT_ROOT}/ca.key ${GIT_ROOT}/ca.csr ${GIT_ROOT}/root-ca.cnf ${GIT_ROOT}/ca.crt

echo
echo "ca-bundle.crt = Use this file to setup proxy instance"
echo "ca.pem = Use this file on Openshift cluster"
echo