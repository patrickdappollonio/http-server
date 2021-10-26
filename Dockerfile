FROM golang:1.17 as builder
WORKDIR /app
ADD . .
ENV CGO_ENABLED=0
RUN go build -trimpath -ldflags='-extldflags=-static -w -s' -o http-server

FROM scratch
COPY html /html
COPY --from=builder /app/http-server /http-server
EXPOSE 5000
ENTRYPOINT ["/http-server"]
