FROM golang:1.13 as builder

WORKDIR /ethereum_exporter
COPY . .

RUN CGO_ENABLED=0 go build -ldflags '-s -w' github.com/31z4/ethereum-prometheus-exporter/cmd/ethereum_exporter

FROM scratch

ENTRYPOINT ["/ethereum_exporter"]
USER nobody
EXPOSE 9368

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /ethereum_exporter/ethereum_exporter /ethereum_exporter
