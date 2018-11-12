FROM golang:1.10
WORKDIR /go/src/k8s.io/minikube
COPY . .
RUN go build -o /gvisor cmd/gvisor/gvisor.go


FROM gcr.io/google-appengine/debian9:latest
RUN apt-get update && apt-get install -y wget systemd curl build-essential gawk
RUN apt-get install sudo
# WORKDIR /src/gnu/
# RUN curl -LO http://ftp.gnu.org/gnu/glibc/glibc-2.26.tar.gz && \
#     tar -xzf glibc-2.26.tar.gz
# WORKDIR /src/gnu/glibc-build
# RUN ../glibc-2.26/configure --disable-sanity-checks
COPY --from=0 /gvisor /gvisor
CMD /gvisor
