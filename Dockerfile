FROM quay.io/prometheus/busybox:latest

COPY mcrouter_exporter /bin/mcrouter_exporter

ENTRYPOINT ["/bin/mcrouter_exporter"]
EXPOSE     9151
