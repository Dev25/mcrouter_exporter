FROM quay.io/prometheus/busybox:latest

COPY mcrouter_exporter_docker /bin/mcrouter_exporter

ENTRYPOINT ["/bin/mcrouter_exporter"]
EXPOSE     9151
