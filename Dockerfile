# sabakan container

# Stage1: build from source
FROM quay.io/cybozu/golang:1.12-bionic AS build

ARG SABAKAN_VERSION=2.2.0

WORKDIR /work

RUN curl -fsSL -o sabakan.tar.gz https://github.com/cybozu-go/sabakan/archive/v${SABAKAN_VERSION}.tar.gz \
    && tar -x -z --strip-components 1 -f sabakan.tar.gz \
    && rm -f sabakan.tar.gz \
    && go get -d ./... \
    && go install ./...

RUN apt-get update \
    && apt-get -y install --no-install-recommends grub-ipxe \
    && rm -rf /var/lib/apt/lists/*

# Stage2: setup runtime container
FROM quay.io/cybozu/ubuntu:18.04

COPY --from=build /go/bin /usr/local/sabakan/bin
COPY --from=build /work/LICENSE /usr/local/sabakan/LICENSE
COPY --from=build /boot/ipxe.efi /usr/lib/ipxe/ipxe.efi
COPY --from=build /usr/share/doc/grub-ipxe/copyright /usr/share/doc/grub-ipxe/copyright
COPY install-tools /usr/local/sabakan/install-tools

VOLUME /var/lib/sabakan
ENV PATH=/usr/local/sabakan/bin:"$PATH"

ENTRYPOINT ["/usr/local/sabakan/bin/sabakan", "-dhcp-bind", "0.0.0.0:67", "-logformat", "json"]
