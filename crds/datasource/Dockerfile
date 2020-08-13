# Build the manager binary
FROM golang:1.13 as builder

WORKDIR /go/src/github.com/chinamobile/nlpt/
# Copy the Go Modules manifests
COPY cmd/ cmd/
COPY apiserver/resources/datasource/rdb/driver/ apiserver/resources/datasource/rdb/driver/
COPY apiserver/resources/datasource/mongo/ apiserver/resources/datasource/mongo/
COPY crds/datasource crds/datasource
COPY pkg/ pkg/
COPY vendor/ vendor/

# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
#RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=off go build -a -o /go/bin/fusion-datasource-controller-manager cmd/datasource-controller-manager/datasource-controller-manager.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /go/bin/fusion-datasource-controller-manager /bin/
USER nonroot:nonroot

ENTRYPOINT ["fusion-datasource-controller-manager"]
