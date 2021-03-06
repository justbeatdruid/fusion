module github.com/chinamobile/nlpt

go 1.12

require (
	github.com/360EntSecGroup-Skylar/excelize v1.4.1
	github.com/apache/pulsar-client-go v0.1.0
	github.com/astaxie/beego v1.12.1
	github.com/coreos/go-systemd v0.0.0-20190321100706-95778dfbb74e
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/docker/docker v0.7.3-0.20190327010347-be7ac8be2ae0
	github.com/fission/fission v1.10.0
	github.com/go-logr/logr v0.1.0
	github.com/go-sql-driver/mysql v1.5.0
	github.com/google/btree v1.0.0
	github.com/hashicorp/go-multierror v1.1.0
	github.com/json-iterator/go v1.1.8
	github.com/lib/pq v1.3.0
	github.com/onsi/ginkgo v1.10.1
	github.com/onsi/gomega v1.7.0
	github.com/parnurzeal/gorequest v0.2.16
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.0.0
	github.com/satori/go.uuid v1.2.0
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5
	github.com/tealeg/xlsx v1.0.5
	go.etcd.io/etcd v0.0.0-20191023171146-3cf2f69b5738
	go.mongodb.org/mongo-driver v1.3.4
	golang.org/x/xerrors v0.0.0-20191204190536-9bdfabe68543 // indirect
	gopkg.in/yaml.v2 v2.2.7 // indirect
	k8s.io/api v0.17.0
	k8s.io/apimachinery v0.17.0
	k8s.io/apiserver v0.17.0
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/component-base v0.17.0
	k8s.io/klog v1.0.0
	moul.io/http2curl v1.0.0 // indirect
	sigs.k8s.io/controller-runtime v0.4.0
)

replace k8s.io/client-go => k8s.io/client-go v0.17.0
