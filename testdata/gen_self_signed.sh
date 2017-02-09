#!/bin/bash

mkdir -p certs && cd certs

openssl genrsa -out ca.key 2048
openssl req -x509 -new -nodes -key ca.key -sha256 -days 3650 -out ca.pem


openssl genrsa -out server.key 2048
openssl req -new -key server.key -out server.csr

# Fix 'cannot validate certificate for 127.0.0.1 because it doesn't contain any IP SANs'
# Refer to http://serverfault.com/questions/611120/failed-tls-handshake-does-not-contain-any-ip-sans, Greg's answer

echo subjectAltName = IP:127.0.0.1 > extfile.cnf
openssl x509 -req -days 3650 -in server.csr -CA ca.pem -CAkey ca.key -CAcreateserial -out server.crt -extfile extfile.cnf

# Remove useless files
rm server.csr extfile.cnf ca.srl
