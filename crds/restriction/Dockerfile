# Build the manager binary
FROM golang:1.13 as builder

WORKDIR /go/src/github.com/chinamobile/nlpt/
# Copy the Go Modules manifests
#COPY go.mod go.mod
#COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
#RUN go mod download

# Copy the go source
# COPY apiserver/ apiserver/
COPY crds/ crds/
COPY pkg/ pkg/
COPY vendor/ vendor/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=off go build -a -o /go/bin/restriction-controller crds/restriction/main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /go/bin/restriction-controller /usr/local/bin
EXPOSE 8080

ENTRYPOINT ["restriction-controller"]
