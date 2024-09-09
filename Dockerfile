FROM busybox as builder
RUN adduser -D -H -u 65534 -G nogroup nobody

FROM scratch
WORKDIR /html
COPY --from=builder /etc/passwd /etc/passwd
USER nobody
COPY http-server /http-server
EXPOSE 5000
ENTRYPOINT ["/http-server"]
