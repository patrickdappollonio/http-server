FROM scratch

ADD html /html
COPY docker-http-server list.tmpl /

EXPOSE 5000
ENTRYPOINT ["./docker-http-server"]
