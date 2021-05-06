module github.com/cybozu-go/sabakan/v2

go 1.16

replace google.golang.org/grpc => google.golang.org/grpc v1.26.0

require (
	github.com/99designs/gqlgen v0.13.0
	github.com/ajeddeloh/go-json v0.0.0-20170920214419-6a2fe990e083 // indirect
	github.com/coreos/etcd v3.3.25+incompatible
	github.com/coreos/ignition v0.35.0
	github.com/cybozu-go/etcdutil v1.3.6
	github.com/cybozu-go/log v1.6.0
	github.com/cybozu-go/netutil v1.3.0
	github.com/cybozu-go/well v1.10.0
	github.com/google/go-cmp v0.5.4
	github.com/google/go-tpm v0.3.2
	github.com/hashicorp/go-version v1.2.1
	github.com/onsi/ginkgo v1.14.2
	github.com/onsi/gomega v1.10.4
	github.com/prometheus/client_golang v1.9.0
	github.com/prometheus/client_model v0.2.0
	github.com/prometheus/common v0.15.0
	github.com/spf13/cobra v1.1.1
	github.com/vektah/gqlparser/v2 v2.1.0
	github.com/vincent-petithory/dataurl v0.0.0-20191104211930-d1553a71de50
	go.universe.tf/netboot v0.0.0-20181010164912-24067fad46fd
	go4.org v0.0.0-20181109185143-00e24f1b2599 // indirect
	golang.org/x/crypto v0.0.0-20201221181555-eec23a3978ad
	sigs.k8s.io/yaml v1.2.0
)
