
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

all: 


clean:
	rm bin/fusion*

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

topic-image: image := ${REG}/library/fusion-topic-controller-manager:0.1.0
topic-image:
	docker build . -f crds/topic/Dockerfile -t ${image}
	docker push ${image}

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	CGO_CFLAGS="-I$(shell pwd)/include" CGO_LDFLAGS="-L$(shell pwd)/lib" go vet ./...

vd:
	go mod tidy
	go mod vendor

alpine:
	docker build -t alpine:glibc -f alpine.Dockerfile .

### following methods for local test
apiserver-binary:
	@if rm bin/fusion-apiserver;then test 0;fi
	CGO_CFLAGS="-I$(shell pwd)/include" CGO_LDFLAGS="-L$(shell pwd)/lib" go build -o bin/fusion-apiserver cmd/apiserver/apiserver.go

datasource-binary:
	@if rm bin/datasource-controller-manager;then test 0;fi
	go build -o bin/fusion-datasource-controller-manager cmd/datasource-controller-manager/datasource-controller-manager.go

application-binary:
	@if rm bin/application-controller-manager;then test 0;fi
	go build -o bin/fusion-application-controller-manager cmd/application-controller-manager/application-controller-manager.go

serviceunit-binary:
	@if rm bin/serviceunit-controller-manager;then test 0;fi
	go build -o bin/fusion-serviceunit-controller-manager cmd/serviceunit-controller-manager/serviceunit-controller-manager.go

api-binary:
	@if rm bin/api-controller-manager;then test 0;fi
	go build -o bin/fusion-api-controller-manager cmd/api-controller-manager/api-controller-manager.go

apply-binary:
	@if rm bin/apply-controller-manager;then test 0;fi
	go build -o bin/fusion-apply-controller-manager cmd/apply-controller-manager/apply-controller-manager.go

apiserver-run:
	LD_LIBRARY_PATH=$(shell pwd)/lib $(shell pwd)/bin/fusion-apiserver --kubeconfig=/root/.kube/config --v=5

datasource-run:
	$(shell pwd)/bin/fusion-datasource-controller-manager --kubeconfig=/root/.kube/config --v=5

application-run:
	$(shell pwd)/bin/fusion-application-controller-manager --kubeconfig=/root/.kube/config --operator-host=119.3.248.187 --operator-port=30081

serviceunit-run:
	$(shell pwd)/bin/fusion-serviceunit-controller-manager --kubeconfig=/root/.kube/config --operator-host=119.3.248.187 --operator-port=30081

api-run:
	$(shell pwd)/bin/fusion-api-controller-manager --kubeconfig=/root/.kube/config --operator-host=119.3.248.187 --operator-port=30081 --portal-port=30080

apply-run:
	$(shell pwd)/bin/fusion-apply-controller-manager --kubeconfig=/root/.kube/config --operator-host=119.3.248.187 --operator-port=30081
