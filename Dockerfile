FROM alpine:3.22.1
RUN apk add ca-certificates ceph-common
COPY ./out/linux-amd64/rbd-exporter /
USER nobody
EXPOSE 9876
ENTRYPOINT ["/rbd-exporter"]
