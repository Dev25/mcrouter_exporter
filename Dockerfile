FROM quay.io/prometheus/busybox:latest

COPY mcrouter_exporter_docker /bin/mcrouter_exporter

CMD ["/bin/mcrouter_exporter"]
EXPOSE     9151
