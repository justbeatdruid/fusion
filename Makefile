
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


# Run tests
test: fmt vet
	go test ./... -coverprofile cover.out

apiserver-binary:
	@if rm bin/fusion-apiserver;then test 0;fi
	CGO_CFLAGS="-I$(shell pwd)/include" CGO_LDFLAGS="-L$(shell pwd)/lib" go build -o bin/fusion-apiserver cmd/apiserver/apiserver.go

apiserver-image: #fmt vet vd
	docker build . -f apiserver.Dockerfile -t ${APIIMG}
	docker push ${APIIMG}

datasource-binary:
	@if rm bin/datasource-controller-manager;then test 0;fi
	go build -o bin/datasource-controller-manager cmd/datasource-controller-manager/datasource-controller-manager.go

datasource-image:
	image=${REG}/library/fusion-datasource-controller-manager:0.1.0
	docker build . -f crds/datasource/Dockerfile -t ${image}
	docker push ${image}

application-binary:
	@if rm bin/application-controller-manager;then test 0;fi
	go build -o bin/application-controller-manager cmd/application-controller-manager/application-controller-manager.go

application-image:
	image=${REG}/library/fusion-application-controller-manager:0.1.0
	docker build . -f crds/application/Dockerfile -t ${image}
	docker push ${image}

serviceunit-binary:
	@if rm bin/serviceunit-controller-manager;then test 0;fi
	go build -o bin/serviceunit-controller-manager cmd/serviceunit-controller-manager/serviceunit-controller-manager.go

serviceunit-image:
	image=${REG}/library/fusion-serviceunit-controller-manager:0.1.0
	docker build . -f crds/serviceunit/Dockerfile -t ${image}
	docker push ${image}

api-binary:
	@if rm bin/api-controller-manager;then test 0;fi
	go build -o bin/api-controller-manager cmd/api-controller-manager/api-controller-manager.go

api-image:
	image=${REG}/library/fusion-api-controller-manager:0.1.0
	docker build . -f crds/api/Dockerfile -t ${image}
	docker push ${image}

apply-binary:
	@if rm bin/apply-controller-manager;then test 0;fi
	go build -o bin/apply-controller-manager cmd/apply-controller-manager/apply-controller-manager.go

apply-image:
	image=${REG}/library/fusion-apply-controller-manager:0.1.0
	docker build . -f crds/apply/Dockerfile -t ${image}
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

apiserver-run:
	LD_LIBRARY_PATH=$(shell pwd)/lib $(shell pwd)/bin/fusion-apiserver --kubeconfig=/root/.kube/config --v=5

datasource-run:
	$(shell pwd)/bin/datasource-controller-manager --kubeconfig=/root/.kube/config --v=5

application-run:
	$(shell pwd)/bin/application-controller-manager --kubeconfig=/root/.kube/config --operator-host=119.3.248.187 --operator-port=30081

serviceunit-run:
	$(shell pwd)/bin/serviceunit-controller-manager --kubeconfig=/root/.kube/config --operator-host=119.3.248.187 --operator-port=30081

api-run:
	$(shell pwd)/bin/api-controller-manager --kubeconfig=/root/.kube/config --operator-host=119.3.248.187 --operator-port=30081 --portal-port=30080

apply-run:
	$(shell pwd)/bin/apply-controller-manager --kubeconfig=/root/.kube/config --operator-host=119.3.248.187 --operator-port=30081
