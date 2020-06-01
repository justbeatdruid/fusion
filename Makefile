# Image URL to use all building/pushing image targets
REG ?= registry.cmcc.com
APITAG ?= library/fusion-apiserver:0.1.0
APIIMG ?= ${REG}/${APITAG}
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

all: build uninstall install

build: clean apiserver-image datasource-image application-image trafficcontrol-image restriction-image topic-image api-image serviceunit-image apply-image topicgroup-image clientauth-image

install:
	kubectl create configmap err-config --from-file=config/err.json && kubectl create -f yaml

uninstall:
	kubectl delete -f yaml; kubectl delete configmaps err-config

crd:
	bash -c 'for i in crds/*;do echo $$i;kubectl apply -f $$i/config/crd/bases;done'

clean:

# Run tests
test: fmt vet
	go test ./... -coverprofile cover.out

apiserver-image: #fmt vet vd
	docker build . -f apiserver.Dockerfile -t ${APIIMG}
	docker push ${APIIMG}

datasource-image: image := ${REG}/library/fusion-datasource-controller-manager:0.1.0
datasource-image:
	docker build . -f crds/datasource/Dockerfile -t ${image}
	docker push ${image}

application-image: applicationimage:= ${REG}/library/fusion-application-controller-manager:0.1.0
application-image:
	docker build . -f crds/application/Dockerfile -t ${applicationimage}
	docker push ${applicationimage}

trafficcontrol-image: trafficcontrolimage:= ${REG}/library/fusion-trafficcontrol-controller-manager:0.1.0
trafficcontrol-image:
	docker build . -f crds/trafficcontrol/Dockerfile -t ${trafficcontrolimage}
	docker push ${trafficcontrolimage}

restriction-image: resimg := ${REG}/library/fusion-restriction-controller-manager:0.1.0
restriction-image:
	docker build . -f crds/restriction/Dockerfile -t ${resimg}
	docker push ${resimg}
	
topic-image: image := ${REG}/library/fusion-topic-controller-manager:0.1.0
topic-image:
	docker build . -f crds/topic/Dockerfile -t ${image}
	docker push ${image}

api-image: apiimg := ${REG}/library/fusion-api-controller-manager:0.1.0
api-image:
	docker build . -f crds/api/Dockerfile -t ${apiimg}
	docker push ${apiimg}
	
serviceunit-image: simg := ${REG}/library/fusion-serviceunit-controller-manager:0.1.0
serviceunit-image:
	docker build . -f crds/serviceunit/Dockerfile -t ${simg}
	docker push ${simg}

apply-image: aplimg := ${REG}/library/fusion-apply-controller-manager:0.1.0
apply-image:
	docker build . -f crds/apply/Dockerfile -t ${aplimg}
	docker push ${aplimg}

topicgroup-image: image := ${REG}/library/fusion-topicgroup-controller-manager:0.1.0
topicgroup-image:
	docker build . -f crds/topicgroup/Dockerfile -t ${image}
	docker push ${image}

clientauth-image: image := ${REG}/library/fusion-clientauth-controller-manager:0.1.0
clientauth-image:
	docker build . -f crds/clientauth/Dockerfile -t ${image}
	docker push ${image}

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet --composites=false ./...

vd:
	go mod tidy
	go mod vendor

alpine:
	docker build -t alpine:glibc -f alpine.Dockerfile .

### following methods for local test
apiserver-binary:
	@if rm bin/fusion-apiserver;then test 0;fi
	go build -o bin/fusion-apiserver cmd/apiserver/apiserver.go

datasource-binary:
	@if rm bin/fusion-datasource-controller-manager;then test 0;fi
	go build -o bin/fusion-datasource-controller-manager cmd/datasource-controller-manager/datasource-controller-manager.go

application-binary:
	@if rm bin/fusion-application-controller-manager;then test 0;fi
	go build -o bin/fusion-application-controller-manager cmd/application-controller-manager/application-controller-manager.go

serviceunit-binary:
	@if rm bin/fusion-serviceunit-controller-manager;then test 0;fi
	go build -o bin/fusion-serviceunit-controller-manager cmd/serviceunit-controller-manager/serviceunit-controller-manager.go

api-binary:
	@if rm bin/fusion-api-controller-manager;then test 0;fi
	go build -o bin/fusion-api-controller-manager cmd/api-controller-manager/api-controller-manager.go

apply-binary:
	@if rm bin/fusion-apply-controller-manager;then test 0;fi
	go build -o bin/fusion-apply-controller-manager cmd/apply-controller-manager/apply-controller-manager.go

apiserver-run:
	$(shell pwd)/bin/fusion-apiserver --kubeconfig=/root/.kube/config --v=5 --dataservice-data-host=10.160.32.24 --audit-host=10.160.32.24 --audit-port=30068 --cas-host=10.160.32.24 --cas-port=30090 --tenant-enabled=true --local-config=$(shell pwd)/config/err.json --tenant-enabled=true --pulsar-admin-token=$(shell pwd)/config/superUserToken --pulsar-token-secret=$(shell pwd)/config/tokenSecret --dataservice-data-host=10.160.32.5 --dataservice-metadata-host=10.160.32.5 --etcd-endpoints=http://10.160.32.24:12379 --database-host=10.160.32.24 --database-port=3306 --database-username=root --database-password=123456 --database-databasename=fusion

apiserver-dataservice-run:
	$(shell pwd)/bin/fusion-apiserver --kubeconfig=/root/.kube/config --v=5 --audit-host=10.160.32.24 --audit-port=30068 --cas-host=10.160.32.24 --cas-port=30090 --tenant-enabled=true --local-config=$(shell pwd)/config/err.json --tenant-enabled=false --pulsar-admin-token=$(shell pwd)/config/superUserToken --pulsar-token-secret=$(shell pwd)/config/tokenSecret --dataservice-data-host=10.160.32.5 --dataservice-metadata-host=10.160.32.5 --etcd-endpoints=http://10.160.32.24:12379 --database-enabled=false

apiserver-sync:
	$(shell pwd)/bin/fusion-apiserver --kubeconfig=/root/.kube/config  --database-host=10.160.32.24 --database-port=3306 --database-username=root --database-password=123456 --database-databasename=fusion --sync-mode=true --v=5 --cas-host=10.160.32.24 --cas-port=30090 --tenant-enabled=true

datasource-run:
	$(shell pwd)/bin/fusion-datasource-controller-manager --kubeconfig=/root/.kube/config --v=5 --dataservice-host=10.160.32.24 --sync-loop-enabled=false

application-run:
	$(shell pwd)/bin/fusion-application-controller-manager --kubeconfig=/root/.kube/config --operator-host=119.3.248.187 --operator-port=30081

serviceunit-run:
	$(shell pwd)/bin/fusion-serviceunit-controller-manager --kubeconfig=/root/.kube/config --operator-host=119.3.248.187 --operator-port=30081

api-run:
	$(shell pwd)/bin/fusion-api-controller-manager --kubeconfig=/root/.kube/config --operator-host=119.3.248.187 --operator-port=30081 --portal-port=30080

apply-run:
	$(shell pwd)/bin/fusion-apply-controller-manager --kubeconfig=/root/.kube/config
