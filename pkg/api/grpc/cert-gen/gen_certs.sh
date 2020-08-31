#!/bin/bash

if [ ! -d "./certs" ]; then
  mkdir $CBMON_ROOT/certs
fi

# Create Root signing Key
openssl genrsa -out $CBMON_ROOT/certs/ca.key 4096

# Generate self-signed Root certificate
openssl req -new -x509 -key $CBMON_ROOT/certs/ca.key -sha256 -subj "/C=KR/ST=DJ/O=Test CA, Inc." -days 3650 -out $CBMON_ROOT/certs/ca.crt

# Create a Key certificate for your service
openssl genrsa -out $CBMON_ROOT/certs/server.key 4096

# Create signing CSR
openssl req -new -key $CBMON_ROOT/certs/server.key -out $CBMON_ROOT/certs/server.csr -config certificate.conf

# Generate a certificate for the service
openssl x509 -req -in $CBMON_ROOT/certs/server.csr -CA $CBMON_ROOT/certs/ca.crt -CAkey $CBMON_ROOT/certs/ca.key -CAcreateserial -out $CBMON_ROOT/certs/server.crt -days 3650 -sha256 -extfile certificate.conf -extensions req_ext