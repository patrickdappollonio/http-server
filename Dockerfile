FROM scratch
COPY http-server /http-server
WORKDIR /html
EXPOSE 5000
ENTRYPOINT ["/http-server"]
