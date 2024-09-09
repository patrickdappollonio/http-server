FROM bash
RUN echo "nobody:x:65534:65534:Nobody:/:" > /etc/nobody

FROM scratch
WORKDIR /html
COPY --from=0 /etc/nobody /etc/passwd
USER nobody
COPY http-server /http-server
EXPOSE 5000
ENTRYPOINT ["/http-server"]
