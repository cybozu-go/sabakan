module github.com/cybozu-go/sabakan/v2

replace google.golang.org/grpc => google.golang.org/grpc v1.26.0

require (
	github.com/99designs/gqlgen v0.9.3
	github.com/agnivade/levenshtein v1.0.2 // indirect
	github.com/ajeddeloh/go-json v0.0.0-20170920214419-6a2fe990e083 // indirect
	github.com/coreos/etcd v3.3.15+incompatible
	github.com/coreos/ignition v0.30.0
	github.com/cybozu-go/etcdutil v1.3.4
	github.com/cybozu-go/log v1.5.0
	github.com/cybozu-go/netutil v1.2.0
	github.com/cybozu-go/well v1.8.1
	github.com/google/go-cmp v0.5.4
	github.com/google/go-tpm v0.3.2
	github.com/hashicorp/go-version v1.0.0
	github.com/hashicorp/golang-lru v0.5.1 // indirect
	github.com/onsi/ginkgo v1.8.0
	github.com/onsi/gomega v1.5.0
	github.com/prometheus/client_golang v0.9.3
	github.com/prometheus/client_model v0.0.0-20190812154241-14fe0d1b01d4
	github.com/prometheus/common v0.4.0
	github.com/spf13/cobra v1.0.0
	github.com/vektah/gqlparser v1.1.2
	github.com/vincent-petithory/dataurl v0.0.0-20160330182126-9a301d65acbb
	go.universe.tf/netboot v0.0.0-20181010164912-24067fad46fd
	go4.org v0.0.0-20181109185143-00e24f1b2599 // indirect
	golang.org/x/crypto v0.0.0-20200221231518-2aa609cf4a9d
	golang.org/x/text v0.3.3 // indirect
	sigs.k8s.io/yaml v1.1.0
)

go 1.13
