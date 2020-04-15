# Build the manager binary
FROM golang:1.13 as builder

WORKDIR /go/src/github.com/chinamobile/nlpt/
# Copy the Go Modules manifests
COPY cmd/ cmd/
COPY crds/clientauth crds/clientauth
COPY pkg/ pkg/
COPY vendor/ vendor/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=off go build -a -o /go/bin/fusion-clientauth-controller-manager cmd/clientauth-controller-manager/clientauth-controller-manager.go


# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /go/bin/fusion-topicgroup-controller-manager /bin/
USER nonroot:nonroot

ENTRYPOINT ["fusion-topicgroup-controller-manager"]
