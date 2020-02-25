# Build the manager binary
FROM golang:1.13 as builder

WORKDIR /go/src/github.com/chinamobile/nlpt/
# Copy the Go Modules manifests
COPY cmd/ cmd/
COPY crds/application crds/application
COPY crds/api crds/api
COPY pkg/ pkg/
COPY vendor/ vendor/
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=off go build -a -o /go/bin/fusion-application-controller-manager cmd/application-controller-manager/application-controller-manager.go

# Copy the go source
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /go/bin/fusion-application-controller-manager /bin/
USER nonroot:nonroot

ENTRYPOINT ["fusion-application-controller-manager"]
