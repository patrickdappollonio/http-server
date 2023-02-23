FROM busybox
RUN echo "nobody:x:65534:65534:Nobody:/:" > /etc/nobody

FROM scratch
USER nobody
WORKDIR /html
COPY http-server /http-server
COPY --from=0 /etc/nobody /etc/passwd
EXPOSE 5000
ENTRYPOINT ["/http-server"]
