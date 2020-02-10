
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
	CGO_CFLAGS="-I$(shell pwd)/include" CGO_LDFLAGS="-L$(shell pwd)/lib" go build -o bin/fusion-apiserver cmd/apiserver/apiserver.go

apiserver-image: #fmt vet vd
	docker build . -f apiserver.Dockerfile -t ${APIIMG}
	docker push ${APIIMG}

serviceunit: fmt vet
	docker build . -f crds/serviceunit/Dockerfile -t ${APIIMG}
	docker push ${APIIMG}

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

run:
	LD_LIBRARY_PATH=$(shell pwd)/lib $(shell pwd)/bin/fusion-apiserver --kubeconfig=/root/.kube/config --v=5
