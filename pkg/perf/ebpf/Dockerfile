# Need an image with chroot
FROM alpine:3
RUN apk -U add ca-certificates
COPY ebpf /ebpf
CMD ["/ebpf"]
