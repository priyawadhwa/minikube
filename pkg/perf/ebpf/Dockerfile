FROM golang:1.13
WORKDIR /build 
COPY . . 
RUN go build -o /ebpf ./pkg/perf/ebpf/*.go

FROM golang:alpine
COPY --from=0 /ebpf /ebpf
CMD /ebpf
