#!/bin/sh

TARGET="$1"

sudo -b sh -c "echo \$\$ >/tmp/placemat_pid$$; exec $PLACEMAT output/cluster.yml" >/dev/null 2>&1
sleep 1
PLACEMAT_PID=$(cat /tmp/placemat_pid$$)
echo "placemat PID: $PLACEMAT_PID"

fin() {
    chmod 600 ./mtest_key
    echo "-------- host1: sabakan log"
    ./mssh ubuntu@${HOST1} sudo journalctl -u sabakan.service --no-pager
    echo "-------- host2: sabakan log"
    ./mssh ubuntu@${HOST2} sudo journalctl -u sabakan.service --no-pager
    echo "-------- host3: sabakan log"
    ./mssh ubuntu@${HOST3} sudo journalctl -u sabakan.service --no-pager
}
trap fin INT TERM HUP 0

$GINKGO -v -focus="${TARGET}" $SUITE_PACKAGE
RET=$?

exit $RET
