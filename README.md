# Ethereum Prometheus Exporter

[![CircleCI](https://circleci.com/gh/31z4/ethereum-prometheus-exporter.svg?style=shield&circle-token=3c4469ca8c3360117a7b843958e5537fa2530682)](https://circleci.com/gh/31z4/ethereum-prometheus-exporter)
[![codecov](https://codecov.io/gh/31z4/ethereum-prometheus-exporter/branch/master/graph/badge.svg)](https://codecov.io/gh/31z4/ethereum-prometheus-exporter)
[![Go Report Card](https://goreportcard.com/badge/github.com/31z4/ethereum-prometheus-exporter)](https://goreportcard.com/report/github.com/31z4/ethereum-prometheus-exporter)

This service exports various metrics from Ethereum clients for consumption by [Prometheus](https://prometheus.io). It uses [JSON-RPC](https://github.com/ethereum/wiki/wiki/JSON-RPC) interface to collect the metrics. Any JSON-RPC 2.0 enabled client should be supported. Although, it has only been tested with [OpenEthereum](https://openethereum.github.io/).

## Usage

You can deploy this exporter using the [31z4/ethereum-prometheus-exporter](https://hub.docker.com/r/31z4/ethereum-prometheus-exporter/) Docker image.

    docker run -d -p 9368:9368 --name ethereum-exporter 31z4/ethereum-prometheus-exporter -url http://ethereum:8545

Keep in mind that your container needs to be able to communicate with the Ethereum client using the specified `url` (default is `http://localhost:8545`).

By default the exporter serves on `:9368` at `/metrics`. The listen address can be changed by specifying the `-addr` flag.

Here is an example [`scrape_config`](https://prometheus.io/docs/prometheus/latest/configuration/configuration/#scrape_config) for Prometheus.

```yaml
- job_name: ethereum
  static_configs:
  - targets:
    - ethereum-exporter:9368
```

## Exported Metrics

| Name | Description |
| ---- | ----------- |
| net_peers | Number of peers currently connected to the client. |
| eth_block_number | Number of the most recent block. |
| eth_block_timestamp | Timestamp of the most recent block. |
| eth_gas_price | Current gas price in wei. *Might be inaccurate*. |
| eth_earliest_block_transactions | Number of transactions in the earliest block. |
| eth_latest_block_transactions | Number of transactions in the latest block. |
| eth_pending_block_transactions | The number of transactions in pending block. |
| eth_hashrate | Hashes per second that this node is mining with. |
| eth_sync_starting | Block number at which current import started. |
| eth_sync_current | Number of most recent block. |
| eth_sync_highest | Estimated number of highest block. |
| parity_net_active_peers | Number of active peers. *Available only for OpenEthereum*. |
| parity_net_connected_peers | Number of peers currently connected to this client. *Available only for OpenEthereum*. |

## Development

[Go modules](https://github.com/golang/go/wiki/Modules) is used for dependency management. Hence Go 1.11 is a minimum required version.

[CircleCI Local CLI](https://circleci.com/docs/2.0/local-cli/) can be used to ensure that everything builds locally.

    circleci build --job lint
    circleci build --job test
    circleci build --job build

## Local setup with Grafana Agent
A local setup has been created to run locally the exporter, with a Grafana Agent instance scraping metrics from it, and pushing them to Grafana Cloud.

**Pre-requisites**
- Docker installed
- Corroborate docker compose is installed with `docker-compose --version`

**Steps**
1. Go to the root directory of the repo.
2. Create a file called `.env`, configuring the following:
```
PROM_REMOTE_WRITE_URL=<Grafana Cloud prometheus remote write URL>
PROM_GCOM_USER_ID=<Prometheus instance id>
PROM_GCOM_API_KEY=<Grafana Cloud API Key with MetricsPublisher role>
```

3. Run: `docker-compose up -d`

## Contributing

Contributions are greatly appreciated. The project follows the typical GitHub pull request model. Before starting any work, please either comment on an existing issue or file a new one.

## Support and Donate

You can show your appreciation for the project and support future development by donating.

[![Donate with Ethereum](https://en.cryptobadges.io/badge/big/0xcaDe516c2c2d916eDf44b958ED5B52C01039fad6)](https://en.cryptobadges.io/donate/0xcaDe516c2c2d916eDf44b958ED5B52C01039fad6)

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
