#!/bin/sh -ex

TARGET=$1

. $(dirname $0)/env
: ${INSTANCE_NAME:?}

# Create GCE instance
$GCLOUD compute instances delete ${INSTANCE_NAME} --zone ${ZONE} || true
$GCLOUD compute instances create ${INSTANCE_NAME} \
  --zone ${ZONE} \
  --machine-type ${MACHINE_TYPE} \
  --image vmx-enabled \
  --boot-disk-type ${DISK_TYPE} \
  --boot-disk-size ${BOOT_DISK_SIZE} \
  --local-ssd interface=scsi \
  --metadata block-project-ssh-keys=TRUE \
  --no-service-account \
  --no-scopes

# Run multi-host test
for i in $(seq 300); do
  if $GCLOUD compute ssh --zone=${ZONE} cybozu@${INSTANCE_NAME} --command=date 2>/dev/null; then
    break
  fi
  sleep 1
done

cat >run.sh <<EOF
#!/bin/sh -e

# mkfs and mount local SSD on /var/scratch
mkfs -t ext4 -F /dev/disk/by-id/google-local-ssd-0
mkdir -p /var/scratch
mount -t ext4 /dev/disk/by-id/google-local-ssd-0 /var/scratch
chmod 1777 /var/scratch

# Run mtest
GOPATH=\$HOME/go
export GOPATH
PATH=/usr/local/go/bin:\$GOPATH/bin:\$PATH
export PATH

# Download Go modules through Takumi Guard.
GOPROXY=https://golang.flatt.tech
export GOPROXY
NETRC=/home/cybozu/.netrc
export NETRC

git clone https://github.com/${GITHUB_REPOSITORY} \
    \$HOME/go/src/github.com/${GITHUB_REPOSITORY}
cd \$HOME/go/src/github.com/${GITHUB_REPOSITORY}
git checkout -qf ${GITHUB_SHA}

cd mtest
cp /assets/etcd-*.tar.gz .
cp /assets/ubuntu-*.img .
make SUDO="sudo" setup
command -v socat >/dev/null || (apt-get update -q && apt-get install -y -q socat) || true
mkdir -p /var/scratch/console
for n in host1 host2 host3 worker1 worker2; do
  ( until socat -u UNIX-CONNECT:/tmp/\$n.socket STDOUT 2>/dev/null; do sleep 0.2; done ) >>/var/scratch/console/\$n.log 2>&1 </dev/null &
  CONSOLE_PIDS="\$CONSOLE_PIDS \$!"
done

make placemat
sleep 3
set +e
make test TARGET="${TARGET}"
RET=\$?
if [ \$RET -ne 0 ]; then
  echo "===== debug: host"
  uname -a
  cat /etc/os-release
  dpkg-query -W 'linux-image-*' qemu-system-x86 swtpm swtpm-tools libtpms0 ipxe-qemu seabios ovmf 2>&1
  ls -l /dev/kvm
  cat /sys/module/kvm_intel/parameters/nested 2>&1
  echo "===== debug: placemat journal"
  journalctl -u placemat.service --no-pager
  echo "===== debug: processes"
  ps -eo pid,etime,stat,args | grep -E '[q]emu|[s]wtpm'
  echo "===== debug: pmctl2"
  pmctl2 node list --json
  echo "===== debug: dmesg"
  dmesg | tail -200
  for f in /var/scratch/console/*.log; do
    echo "===== debug: console \$f"
    tail -c 30000 \$f | tr -cd '\11\12\40-\176'
  done
fi
kill \$CONSOLE_PIDS 2>/dev/null
pkill -f 'socat -u UNIX-CONNECT:/tmp/'
exit \$RET
EOF
chmod +x run.sh

NETRC_FILE=$(mktemp)
if grep -q '^machine golang\.flatt\.tech ' "$HOME/.netrc" 2>/dev/null; then
  grep '^machine golang\.flatt\.tech ' "$HOME/.netrc" > ${NETRC_FILE}
  $GCLOUD compute scp --zone=${ZONE} ${NETRC_FILE} cybozu@${INSTANCE_NAME}:.netrc
fi
rm -f ${NETRC_FILE}

$GCLOUD compute scp --zone=${ZONE} run.sh cybozu@${INSTANCE_NAME}:
set +e
$GCLOUD compute ssh --zone=${ZONE} cybozu@${INSTANCE_NAME} --command='sudo /home/cybozu/run.sh'
RET=$?

exit $RET
