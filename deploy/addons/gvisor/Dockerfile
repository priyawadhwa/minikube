# FROM golang:1.10
# WORKDIR /go/src/k8s.io/minikube
# COPY . .
# RUN go build -o /gvisor cmd/gvisor/gvisor.go


FROM ubuntu:18.04
RUN apt-get update && \
    apt-get install -y kmod gcc wget xz-utils libc6-dev bc libelf-dev bison flex openssl libssl-dev libidn2-0 sudo libcap2 && \
    rm -rf /var/lib/apt/lists/*
COPY out/gvisor /gvisor
CMD /gvisor
