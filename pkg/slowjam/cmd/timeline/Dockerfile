# docker build -t gcr.io/priya-wadhwa/slowjam -f pkg/slowjam/cmd/timeline/Dockerfile .

FROM golang:1.13
WORKDIR /minikube
COPY . .
WORKDIR /minikube/pkg/slowjam/cmd/timeline
RUN go build main.go


FROM golang:1.13
COPY --from=0 /minikube/pkg/slowjam/cmd/timeline/* ./
CMD ./main ./stack.log
EXPOSE 8000
