FROM scratch
COPY proxy-to-host /usr/bin/proxy-to-host
EXPOSE 80
ENTRYPOINT ["/usr/bin/proxy-to-host"]
