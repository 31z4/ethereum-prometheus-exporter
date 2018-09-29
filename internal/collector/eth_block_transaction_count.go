package collector

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/prometheus/client_golang/prometheus"
)

type EthBlockTransactionCount struct {
	rpc  *rpc.Client
	desc *prometheus.Desc
}

func (collector *EthBlockTransactionCount) collectByTag(tag string, ch chan<- prometheus.Metric) {
	var result hexutil.Uint64
	if err := collector.rpc.Call(&result, "eth_getBlockTransactionCountByNumber", tag); err != nil {
		ch <- prometheus.NewInvalidMetric(collector.desc, err)
		return
	}

	value := float64(result)
	ch <- prometheus.MustNewConstMetric(collector.desc, prometheus.GaugeValue, value, tag)
}

func NewEthBlockTransactionCount(rpc *rpc.Client) *EthBlockTransactionCount {
	return &EthBlockTransactionCount{
		rpc: rpc,
		desc: prometheus.NewDesc(
			"eth_block_transaction_count",
			"the number of transactions in a block",
			[]string{"tag"},
			nil,
		),
	}
}

func (collector *EthBlockTransactionCount) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.desc
}

func (collector *EthBlockTransactionCount) Collect(ch chan<- prometheus.Metric) {
	go collector.collectByTag("earliest", ch)
	go collector.collectByTag("latest", ch)
	go collector.collectByTag("pending", ch)
}
