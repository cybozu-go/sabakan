#!/bin/sh
cfssl gencert --initca=true ./ca-csr.json | cfssljson --bare ./ca
mv ca.pem ca.crt
cfssl gencert --ca ./ca.crt  --ca-key ./ca-key.pem --config ./config.json  server-csr.json | cfssljson --bare ./server
mv server.pem server.crt
mv server-key.pem server.key.insecure
