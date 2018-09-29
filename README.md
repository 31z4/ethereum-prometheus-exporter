# Ethereum Prometheus Exporter

[![CircleCI](https://circleci.com/gh/31z4/ethereum-prometheus-exporter.svg?style=shield&circle-token=3c4469ca8c3360117a7b843958e5537fa2530682)](https://circleci.com/gh/31z4/ethereum-prometheus-exporter)
[![Go Report Card](https://goreportcard.com/badge/github.com/31z4/ethereum-prometheus-exporter)](https://goreportcard.com/report/github.com/31z4/ethereum-prometheus-exporter)

***Disclaimer: this project is a work in progress.***

## Exported Metrics

| Name | Description | Labels |
| ---- | ----------- | ------ |
| net_peer_count | The number of peers currently connected to the client. | |
| eth_block_number | The number of most recent block. | |
| eth_gas_price | The current price per gas in wei. *Might be inaccurate*. | |
| eth_block_transaction_count | The number of transactions in a block. | tag |
| eth_hashrate | The number of hashes per second that the node is mining with. | |
| eth_syncing | Data about the sync status. | block |
| parity_net_peers | The number of peers currently connected to the client. *Available only for Parity*. | status |

## Development

[Go modules](https://golang.org/doc/go1.11#modules) is used for dependency management. Hence Go 1.11 is a minimum required version.

[CircleCI Local CLI](https://circleci.com/docs/2.0/local-cli/) can be used to ensure that everything builds locally.

    circleci build --job lint
    circleci build --job test
    circleci build --job build

## Contributing

Contributions are greatly appreciated. The project follows the typical GitHub pull request model. Before starting any work, please either comment on an existing issue or file a new one.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
