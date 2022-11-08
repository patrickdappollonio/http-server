FROM scratch
COPY http-server /http-server
EXPOSE 5000
ENTRYPOINT ["/http-server"]
