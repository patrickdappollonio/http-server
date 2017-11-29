FROM drone/ca-certs

ADD html /html
ADD http-server /

EXPOSE 5000
ENTRYPOINT ["./http-server"]
