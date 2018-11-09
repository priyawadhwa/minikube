FROM golang:1.10
WORKDIR /go/src/k8s.io/minikube
COPY . .
RUN go build -o /gvisor cmd/gvisor/gvisor.go


FROM gcr.io/google-appengine/debian9:latest
RUN apt-get update && apt-get install -y wget systemd
COPY --from=0 /gvisor /gvisor
CMD /gvisor
