# FROM golang:1.10
# WORKDIR /go/src/k8s.io/minikube
# COPY . .
# RUN go build -o /gvisor cmd/gvisor/gvisor.go


FROM ubuntu:18.04
RUN apt-get update && \
    apt-get install -y kmod gcc wget xz-utils libc6-dev bc libelf-dev bison flex openssl libssl-dev && \
    rm -rf /var/lib/apt/lists/*
RUN apt-get update && apt-get -y install libidn2-0
# RUN apt-get update && apt-get install -y curl
COPY out/gvisor /gvisor
CMD /gvisor
