FROM alpine
EXPOSE 5000

ADD html /html
ADD docker-http-server list.tmpl /usr/local/bin/

ENTRYPOINT ["docker-http-server"]
