#!/bin/sh

if [ "$0" != "./gencerts.sh" ]; then
  echo "must be run from 'testdata'"
  exit 255
fi

if ! which cfssl; then
  echo "cfssl is not installed"
  exit 255
fi

mkdir -p ../output/certs/
cfssl gencert --initca=true ./ca-csr.json | cfssljson --bare ../output/certs/ca
mv ../output/certs/ca.pem ../output/certs/ca.crt
if which openssl >/dev/null; then
  openssl x509 -in ca.crt -noout -text
fi

# generate DNS: localhost, IP: 127.0.0.1, CN: example.com certificates
cfssl gencert \
  --ca ../output/certs//ca.crt \
  --ca-key ../output/certs//ca-key.pem \
  --config ./gencert.json \
  ./server-ca-csr.json | cfssljson --bare ../output/certs/server
mv ../output/certs/server.pem ../output/certs/server.crt
mv ../output/certs/server-key.pem ../output/certs/server.key.insecure

cd ../output/certs/ && rm -f *.pem *.stderr *.txt
