# Sabakan container
FROM quay.io/cybozu/ubuntu:18.04

RUN apt-get update \
    && apt-get -y install --no-install-recommends grub-ipxe \
    && rm -rf /var/lib/apt/lists/* \
    && mkdir -p /usr/lib/ipxe \
    && cp /boot/ipxe.efi /usr/lib/ipxe/ipxe.efi

COPY sabakan /usr/local/sabakan/bin/
COPY sabactl /usr/local/sabakan/bin/
COPY install-tools /usr/local/sabakan/install-tools
COPY LICENSE /usr/local/sabakan/

VOLUME /var/lib/sabakan
ENV PATH=/usr/local/sabakan/bin:"$PATH"

ENTRYPOINT ["/usr/local/sabakan/bin/sabakan", "-dhcp-bind", "0.0.0.0:67", "-logformat", "json"]
