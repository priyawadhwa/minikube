# Need an image with chroot
FROM debian
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates lz4
COPY ebpf /ebpf
CMD ["/ebpf"]
