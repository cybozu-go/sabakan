# Sabakan container
FROM ghcr.io/cybozu/ubuntu:22.04
LABEL org.opencontainers.image.source="https://github.com/cybozu-go/sabakan"

RUN apt-get update \
    && apt-get -y install --no-install-recommends grub-ipxe \
    && rm -rf /var/lib/apt/lists/* \
    && mkdir -p /usr/lib/ipxe \
    && cp /boot/ipxe.efi /usr/lib/ipxe/ipxe.efi

COPY sabakan /usr/local/sabakan/bin/sabakan
COPY sabactl /usr/local/sabakan/bin/sabactl
COPY sabakan-cryptsetup /usr/local/sabakan/bin/sabakan-cryptsetup
COPY install-tools /usr/local/sabakan/install-tools
COPY LICENSE /usr/local/sabakan/LICENSE
RUN chmod -R +xr /usr/local/sabakan/bin

VOLUME /var/lib/sabakan
ENV PATH=/usr/local/sabakan/bin:"$PATH"

ENTRYPOINT ["/usr/local/sabakan/bin/sabakan", "-dhcp-bind", "0.0.0.0:67", "-logformat", "json"]
