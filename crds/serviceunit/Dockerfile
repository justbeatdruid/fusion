# Build the manager binary
FROM golang:1.13 as builder

WORKDIR /go/src/github.com/chinamobile/nlpt/
# Copy the Go Modules manifests
COPY cmd/ cmd/
COPY crds/serviceunit crds/serviceunit
COPY crds/api crds/api
COPY crds/datasource crds/datasource
COPY pkg/ pkg/
COPY vendor/ vendor/

# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
#RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=off go build -a -o /go/bin/fusion-serviceunit-controller-manager cmd/serviceunit-controller-manager/serviceunit-controller-manager.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM alpine:latest
WORKDIR /
COPY --from=builder /go/bin/fusion-serviceunit-controller-manager /bin/
USER nonroot:nonroot

ENTRYPOINT ["fusion-serviceunit-controller-manager"]
