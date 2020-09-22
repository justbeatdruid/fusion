FROM golang:1.13 as builder

WORKDIR /go/src/github.com/chinamobile/oracle/
# Copy the Go Modules manifests
#COPY go.mod go.mod
#COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
#RUN go mod download

# Copy the go source

RUN apt update && apt install unzip build-essential libaio-dev -y

COPY main.go main.go
COPY vendor/ vendor/

COPY instantclient/ instantclient/
RUN cd instantclient && bash install.sh && cd ..

RUN cp /usr/lib/instantclient_12_2/sdk/include/*.h /usr/include

ENV LD_LIBRARY_PATH=/usr/lib/instantclient_12_2
ENV PKG_CONFIG_PATH=/usr/lib/pkg-config

RUN CGO_ENABLED=1 go build -o main main.go

FROM ubuntu:18.04

WORKDIR /

RUN apt update && apt install unzip libaio-dev -y
COPY instantclient/instantclient-basic-linux.x64-12.2.0.1.0.zip /instantclient-basic-linux.x64-12.2.0.1.0.zip
RUN unzip instantclient-basic-linux.x64-12.2.0.1.0.zip

ENV LD_LIBRARY_PATH=/instantclient_12_2
COPY --from=builder /go/src/github.com/chinamobile/oracle/main /main

ENTRYPOINT ["/main"]
