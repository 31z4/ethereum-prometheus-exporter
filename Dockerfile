FROM golang:1.17 as builder

WORKDIR /ethereum_exporter
COPY . .

ARG VERSION=undefined
RUN CGO_ENABLED=0 \
    go build -ldflags "-s -w -X main.version=$VERSION" github.com/thepalbi/ethereum-prometheus-exporter/cmd/ethereum_exporter

FROM scratch

ENTRYPOINT ["/ethereum_exporter"]
USER nobody
EXPOSE 9368

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /ethereum_exporter/ethereum_exporter /ethereum_exporter
