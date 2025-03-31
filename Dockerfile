FROM alpine:3.21.3
RUN apk add ca-certificates ceph-common
COPY ./out/linux-amd64/rbd-exporter /
USER nobody
ENTRYPOINT ["/rbd-exporter"]