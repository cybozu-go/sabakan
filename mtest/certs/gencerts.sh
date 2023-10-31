#!/bin/sh
mkdir -p ../output/certs
cfssl gencert --initca=true ./ca-csr.json | cfssljson --bare ../output/certs/ca
mv ../output/certs/ca.pem ../output/certs/ca.crt
cfssl gencert --ca ../output/certs/ca.crt  --ca-key ../output/certs/ca-key.pem --config ./config.json  server-csr.json | cfssljson --bare ../output/certs/server
mv ../output/certs/server.pem ../output/certs/server.crt
mv ../output/certs/server-key.pem ../output/certs/server.key.insecure
