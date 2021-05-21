module github.com/cybozu-go/sabakan/v2

go 1.16

replace google.golang.org/grpc => google.golang.org/grpc v1.26.0

require (
	github.com/99designs/gqlgen v0.13.0
	github.com/ajeddeloh/go-json v0.0.0-20170920214419-6a2fe990e083 // indirect
	github.com/coreos/ignition v0.35.0
	github.com/cybozu-go/etcdutil v1.4.0
	github.com/cybozu-go/log v1.6.0
	github.com/cybozu-go/netutil v1.4.1
	github.com/cybozu-go/well v1.10.0
	github.com/google/go-cmp v0.5.5
	github.com/google/go-tpm v0.3.2
	github.com/hashicorp/go-version v1.3.0
	github.com/onsi/ginkgo v1.16.2
	github.com/onsi/gomega v1.12.0
	github.com/prometheus/client_golang v1.10.0
	github.com/prometheus/client_model v0.2.0
	github.com/prometheus/common v0.25.0
	github.com/spf13/cobra v1.1.3
	github.com/vektah/gqlparser/v2 v2.2.0
	github.com/vincent-petithory/dataurl v0.0.0-20191104211930-d1553a71de50
	go.etcd.io/etcd v0.5.0-alpha.5.0.20210512015243-d19fbe541bf9
	go.universe.tf/netboot v0.0.0-20201124111825-bdaec9d82638
	go4.org v0.0.0-20181109185143-00e24f1b2599 // indirect
	golang.org/x/crypto v0.0.0-20210513164829-c07d793c2f9a
	sigs.k8s.io/yaml v1.2.0
)
