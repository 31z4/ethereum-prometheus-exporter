package collector

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/prometheus/client_golang/prometheus"
)

type EthLatestBlockTransactions struct {
	rpc  *rpc.Client
	desc *prometheus.Desc
}

func NewEthLatestBlockTransactions(rpc *rpc.Client) *EthLatestBlockTransactions {
	return &EthLatestBlockTransactions{
		rpc: rpc,
		desc: prometheus.NewDesc(
			"eth_latest_block_transactions",
			"the number of transactions in a latest block",
			nil,
			nil,
		),
	}
}

func (collector *EthLatestBlockTransactions) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.desc
}

func (collector *EthLatestBlockTransactions) Collect(ch chan<- prometheus.Metric) {
	var result hexutil.Uint64
	if err := collector.rpc.Call(&result, "eth_getBlockTransactionCountByNumber", "latest"); err != nil {
		ch <- prometheus.NewInvalidMetric(collector.desc, err)
		return
	}

	value := float64(result)
	ch <- prometheus.MustNewConstMetric(collector.desc, prometheus.GaugeValue, value)
}
