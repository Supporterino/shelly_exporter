FROM scratch
ENTRYPOINT ["/shelly_exporter"]
COPY shelly_exporter /