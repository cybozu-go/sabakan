package mtest

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cybozu-go/well"
	. "github.com/onsi/gomega"
	"golang.org/x/crypto/ssh"
)

const (
	sshTimeout         = 3 * time.Minute
	defaultDialTimeout = 30 * time.Second
	defaultKeepAlive   = 5 * time.Second

	// DefaultRunTimeout is the timeout value for Agent.Run().
	DefaultRunTimeout = 10 * time.Minute
)

var (
	sshClients  = make(map[string]*sshAgent)
	agentDialer = &net.Dialer{
		Timeout:   defaultDialTimeout,
		KeepAlive: defaultKeepAlive,
	}
)

type sshAgent struct {
	client *ssh.Client
	conn   net.Conn
}

func sshTo(address string, sshKey ssh.Signer, userName string) (*sshAgent, error) {
	conn, err := agentDialer.Dial("tcp", address+":22")
	if err != nil {
		fmt.Printf("failed to dial: %s\n", address)
		return nil, err
	}
	config := &ssh.ClientConfig{
		User: userName,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(sshKey),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	}
	err = conn.SetDeadline(time.Now().Add(defaultDialTimeout))
	if err != nil {
		conn.Close()
		return nil, err
	}
	clientConn, channelCh, reqCh, err := ssh.NewClientConn(conn, "tcp", config)
	if err != nil {
		// conn was already closed in ssh.NewClientConn
		return nil, err
	}
	err = conn.SetDeadline(time.Time{})
	if err != nil {
		clientConn.Close()
		return nil, err
	}
	a := sshAgent{
		client: ssh.NewClient(clientConn, channelCh, reqCh),
		conn:   conn,
	}
	return &a, nil
}

func prepareSSHClients(addresses ...string) error {
	sshKey, err := parsePrivateKey(sshKeyFile)
	if err != nil {
		return err
	}

	ch := time.After(sshTimeout)
	for _, a := range addresses {
	RETRY:
		select {
		case <-ch:
			return errors.New("prepareSSHClients timed out")
		default:
		}
		agent, err := sshTo(a, sshKey, "cybozu")
		if err != nil {
			time.Sleep(time.Second)
			goto RETRY
		}
		sshClients[a] = agent
	}

	return nil
}

func parsePrivateKey(keyPath string) (ssh.Signer, error) {
	f, err := os.Open(keyPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	return ssh.ParsePrivateKey(data)
}

func loadImage(path string) error {
	env := well.NewEnvironment(context.Background())
	for _, host := range []string{host1, host2, host3} {
		host2 := host
		env.Go(func(ctx context.Context) error {
			sess, err := sshClients[host2].client.NewSession()
			if err != nil {
				return err
			}
			defer func() {
				sess.Run("rm -f " + path)
				sess.Close()
			}()
			return sess.Run("docker load -i " + path)
		})
	}
	env.Stop()
	return env.Wait()

}

func installTools(image string) error {
	env := well.NewEnvironment(context.Background())
	for _, host := range []string{host1, host2, host3} {
		host2 := host
		env.Go(func(ctx context.Context) error {
			sess, err := sshClients[host2].client.NewSession()
			if err != nil {
				return err
			}
			defer sess.Close()
			return sess.Run("docker run --rm -u root:root --entrypoint /usr/local/sabakan/install-tools --mount type=bind,src=/opt/bin,target=/host/usr/local/bin " + image)
		})
	}
	env.Stop()
	return env.Wait()
}

func stopEtcd() error {
	env := well.NewEnvironment(context.Background())
	for _, host := range []string{host1, host2, host3} {
		host2 := host
		env.Go(func(ctx context.Context) error {
			sess, err := sshClients[host2].client.NewSession()
			if err != nil {
				return err
			}
			defer sess.Close()
			return sess.Run("sudo systemctl stop my-etcd.service; sudo rm -rf /home/cybozu/default.etcd")
		})
	}
	env.Stop()
	return env.Wait()
}

func runEtcd() error {
	env := well.NewEnvironment(context.Background())
	for _, host := range []string{host1, host2, host3} {
		host2 := host
		env.Go(func(ctx context.Context) error {
			sess, err := sshClients[host2].client.NewSession()
			if err != nil {
				return err
			}
			defer sess.Close()
			return sess.Run("sudo systemd-run --unit=my-etcd.service /opt/bin/etcd --listen-client-urls=http://0.0.0.0:2379 --advertise-client-urls=http://localhost:2379 --data-dir /home/cybozu/default.etcd")
		})
	}
	env.Stop()
	return env.Wait()
}

func stopSabakan() error {
	env := well.NewEnvironment(context.Background())
	for _, host := range []string{host1, host2, host3} {
		host2 := host
		env.Go(func(ctx context.Context) error {
			sess, err := sshClients[host2].client.NewSession()
			if err != nil {
				return err
			}
			defer sess.Close()
			return sess.Run("sudo systemctl reset-failed sabakan.service; sudo systemctl stop sabakan.service; sudo rm -rf /var/lib/sabakan")
		})
	}
	env.Stop()
	return env.Wait()
}

func runSabakan(image string) error {
	env := well.NewEnvironment(context.Background())
	for _, host := range []string{host1, host2, host3} {
		host2 := host
		env.Go(func(ctx context.Context) error {
			sess, err := sshClients[host2].client.NewSession()
			if err != nil {
				return err
			}
			defer sess.Close()
			return sess.Run("sudo mkdir -p /var/lib/sabakan && sudo systemd-run --unit=sabakan.service docker run --rm --read-only --cap-drop ALL --cap-add NET_BIND_SERVICE --network host --name sabakan --mount type=bind,source=/var/lib/sabakan,target=/var/lib/sabakan --mount type=bind,source=/etc/sabakan,target=/etc/sabakan --mount type=bind,source=/etc/sabakan.yml,target=/etc/sabakan.yml " + image + " -config-file /etc/sabakan.yml")
		})
	}
	env.Stop()
	return env.Wait()
}

func execAt(host string, args ...string) (stdout, stderr []byte, e error) {
	return execAtWithStream(host, nil, args...)
}

// WARNING: `input` can contain secret data.  Never output `input` to console.
func execAtWithStream(host string, input io.Reader, args ...string) (stdout, stderr []byte, e error) {
	agent := sshClients[host]
	return doExec(agent, input, args...)
}

// WARNING: `input` can contain secret data.  Never output `input` to console.
func doExec(agent *sshAgent, input io.Reader, args ...string) ([]byte, []byte, error) {
	err := agent.conn.SetDeadline(time.Now().Add(DefaultRunTimeout))
	if err != nil {
		return nil, nil, err
	}
	defer agent.conn.SetDeadline(time.Time{})

	sess, err := agent.client.NewSession()
	if err != nil {
		return nil, nil, err
	}
	defer sess.Close()

	if input != nil {
		sess.Stdin = input
	}
	outBuf := new(bytes.Buffer)
	errBuf := new(bytes.Buffer)
	sess.Stdout = outBuf
	sess.Stderr = errBuf
	err = sess.Run(strings.Join(args, " "))
	return outBuf.Bytes(), errBuf.Bytes(), err
}

func execSafeAt(host string, args ...string) []byte {
	stdout, stderr, err := execAt(host, args...)
	ExpectWithOffset(1, err).To(Succeed(), "[%s] %v: %s", host, args, stderr)
	return stdout
}

func remoteTempFile(body string) string {
	f, err := os.CreateTemp("", "mtest")
	Expect(err).NotTo(HaveOccurred())
	defer f.Close()
	_, err = f.WriteString(body)
	Expect(err).NotTo(HaveOccurred())
	_, err = f.Seek(0, io.SeekStart)
	Expect(err).NotTo(HaveOccurred())
	remoteFile := filepath.Join("/tmp", filepath.Base(f.Name()))
	_, _, err = execAtWithStream(host1, f, "dd", "of="+f.Name())
	Expect(err).NotTo(HaveOccurred())
	return remoteFile
}

func sabactl(args ...string) ([]byte, []byte, error) {
	args = append([]string{"/opt/bin/sabactl", "--server", "http://" + host1 + ":10080"}, args...)
	return execAt(host1, args...)
}

func sabactlSafe(args ...string) []byte {
	args = append([]string{"/opt/bin/sabactl", "--server", "http://" + host1 + ":10080"}, args...)
	return execSafeAt(host1, args...)
}

func etcdctl(args ...string) ([]byte, []byte, error) {
	args = append([]string{"/opt/bin/etcdctl", "--endpoints=http://" + host1 + ":2379"}, args...)
	return execAt(host1, args...)
}

func stopHost2Sabakan() ([]byte, []byte, error) {
	return execAt(host2, "sudo", "systemctl", "stop", "sabakan.service")
}

func startHost2Sabakan(image string) ([]byte, []byte, error) {
	return execAt(host2, "sudo", "systemd-run", "--unit=sabakan.service",
		"docker", "run", "--rm", "--read-only", "--cap-drop", "ALL", "--cap-add", "NET_BIND_SERVICE",
		"--network", "host", "--name", "sabakan",
		"--mount", "type=bind,source=/var/lib/sabakan,target=/var/lib/sabakan",
		"--mount", "type=bind,source=/etc/sabakan.yml,target=/etc/sabakan.yml",
		"--mount", "type=bind,source=/etc/sabakan,target=/etc/sabakan",
		image, "-config-file=/etc/sabakan.yml")
}
