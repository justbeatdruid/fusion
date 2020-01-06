
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

all: apiserver

# Run tests
test: fmt vet
	go test ./... -coverprofile cover.out

apiserver: fmt vet
	go build -o bin/apiserver apiserver/main.go
	#docker build . -f apiserver/Dockerfile -t ${APIIMG}
	#docker push ${APIIMG}

application-controller:
	go build -o bin/application-controller application-controller/main.go

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

vendor:
	go mod vendor
