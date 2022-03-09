# Ethereum Prometheus Exporter

This service exports various metrics from Ethereum clients for consumption by [Prometheus](https://prometheus.io). It uses [JSON-RPC](https://github.com/ethereum/wiki/wiki/JSON-RPC) interface to collect the metrics. Any JSON-RPC 2.0 enabled client should be supported.

## Exported Metrics

| Name                            | Description                                        |
| ------------------------------- | -------------------------------------------------- |
| net_peers                       | Number of peers currently connected to the client. |
| eth_block_number                | Number of the most recent block.                   |
| eth_block_timestamp             | Timestamp of the most recent block.                |
| eth_gas_price                   | Current gas price in wei. _Might be inaccurate_.   |
| eth_earliest_block_transactions | Number of transactions in the earliest block.      |
| eth_latest_block_transactions   | Number of transactions in the latest block.        |
| eth_pending_block_transactions  | The number of transactions in pending block.       |
| eth_hashrate                    | Hashes per second that this node is mining with.   |
| eth_sync_starting               | Block number at which current import started.      |
| eth_sync_current                | Number of most recent block.                       |
| eth_sync_highest                | Estimated number of highest block.                 |

## Development

[Go modules](https://github.com/golang/go/wiki/Modules) is used for dependency management. Hence Go 1.11 is a minimum required version.

## Local setup with Grafana Agent

A local setup has been created to run locally the exporter, with a Grafana Agent instance scraping metrics from it, and pushing them to Grafana Cloud.

**Pre-requisites**

- Docker installed
- Corroborate docker compose is installed with `docker-compose --version`

**Steps**

1. Go to the root directory of the repo.
2. Create an exporter config file inside the `production/exporter` folder. You can use `sample.yaml` as template.
3. Create a file called `.env`, configuring the following:

```
EXPORTER_CONFIG_FILE=<Exporter config file located in the production/exporter folder, like `sample.yaml`>
PROM_REMOTE_WRITE_URL=<Grafana Cloud prometheus remote write URL>
PROM_GCOM_USER_ID=<Prometheus instance id>
PROM_GCOM_API_KEY=<Grafana Cloud API Key with MetricsPublisher role>
```

4. Run: `docker-compose up -d`

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
