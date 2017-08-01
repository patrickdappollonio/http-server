FROM alpine
EXPOSE 5000

ADD html /html
COPY docker-http-server list.tmpl /usr/local/bin/

ENTRYPOINT ["docker-http-server"]
