FROM busybox:stable-glibc as builder
RUN echo "nobody:x:65534:65534:Nobody:/:" > /etc/nobody

FROM scratch
WORKDIR /html
COPY --from=builder /etc/nobody /etc/passwd
USER nobody
COPY http-server /http-server
EXPOSE 5000
ENTRYPOINT ["/http-server"]
