FROM golang:1.22-alpine as builder
WORKDIR /tcp_proxy
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -buildvcs=false -ldflags="-w -s"

FROM scratch
COPY --from=builder /tcp_proxy/tcp_proxy /
CMD ["/tcp_proxy"]