FROM scratch

ADD html /html
COPY http-server list.tmpl /

EXPOSE 5000
ENTRYPOINT ["./http-server"]
