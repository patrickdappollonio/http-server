FROM scratch
EXPOSE 5000

ADD html /html
COPY docker-http-server list.tmpl /

ENTRYPOINT ["./docker-http-server"]
