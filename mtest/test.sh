#!/bin/sh

TARGET="$1"

fin() {
    chmod 600 ./mtest_key
    echo "-------- host1: sabakan log"
    ./mssh cybozu@${HOST1} sudo journalctl -u sabakan.service --no-pager
    echo "-------- host2: sabakan log"
    ./mssh cybozu@${HOST2} sudo journalctl -u sabakan.service --no-pager
    echo "-------- host3: sabakan log"
    ./mssh cybozu@${HOST3} sudo journalctl -u sabakan.service --no-pager
    echo "-------- worker1: sabakan-cryptsetup log"
    ./mssh cybozu@${WORKER1} sudo journalctl -u sabakan-cryptsetup.service --no-pager
    echo "-------- worker2: sabakan-cryptsetup log"
    ./mssh cybozu@${WORKER2} sudo journalctl -u sabakan-cryptsetup.service --no-pager
}
trap fin INT TERM HUP 0

$GINKGO -v -focus="${TARGET}" $SUITE_PACKAGE
RET=$?

exit $RET
