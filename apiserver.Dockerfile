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
COPY cmd/ cmd/
COPY apiserver/ apiserver/
COPY crds/ crds/
COPY pkg/ pkg/
COPY vendor/ vendor/
COPY lib/libpulsar* /usr/lib/
COPY include/pulsar /usr/include/pulsar

# Build
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 GO111MODULE=off go build -a -o /go/bin/fusion-apiserver cmd/apiserver/apiserver.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM alpine:latest

RUN apk --no-cache add ca-certificates wget
RUN wget https://github.com/sgerrand/alpine-pkg-glibc/releases/download/2.30-r0/glibc-2.30-r0.apk && \
    wget https://github.com/sgerrand/alpine-pkg-glibc/releases/download/2.30-r0/glibc-bin-2.30-r0.apk && \
    wget https://github.com/sgerrand/alpine-pkg-glibc/releases/download/2.30-r0/glibc-i18n-2.30-r0.apk && \
    apk add glibc-2.30-r0.apk glibc-bin-2.30-r0.apk glibc-i18n-2.30-r0.apk && \
    rm glibc-2.30-r0.apk glibc-bin-2.30-r0.apk glibc-i18n-2.30-r0.apk

COPY --from=builder /go/bin/fusion-apiserver /usr/local/bin
COPY lib/libpulsar.so.2.5.0 /lib/
RUN cd /lib && ln -s libpulsar.so.2.5.0 libpulsar.so

EXPOSE 8001

ENTRYPOINT ["fusion-apiserver"]
