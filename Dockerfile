FROM alpine
EXPOSE 5000

ADD html /html
ADD docker-http-server /usr/local/bin/

ENTRYPOINT ["docker-http-server"]
