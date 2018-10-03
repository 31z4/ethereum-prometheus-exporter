package collector

import (
	"sync"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/prometheus/client_golang/prometheus"
)

type EthBlockTransactionCount struct {
	rpc          *rpc.Client
	earliestDesc *prometheus.Desc
	latestDesc   *prometheus.Desc
	pendingDesc  *prometheus.Desc
}

func (collector *EthBlockTransactionCount) collectByTag(tag string, desc *prometheus.Desc, ch chan<- prometheus.Metric) {
	var result hexutil.Uint64
	if err := collector.rpc.Call(&result, "eth_getBlockTransactionCountByNumber", tag); err != nil {
		ch <- prometheus.NewInvalidMetric(desc, err)
		return
	}

	value := float64(result)
	ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, value)
}

func NewEthBlockTransactionCount(rpc *rpc.Client) *EthBlockTransactionCount {
	return &EthBlockTransactionCount{
		rpc: rpc,
		earliestDesc: prometheus.NewDesc(
			"eth_earliest_block_transactions",
			"the number of transactions in an earliest block",
			nil,
			nil,
		),
		latestDesc: prometheus.NewDesc(
			"eth_latest_block_transactions",
			"the number of transactions in a latest block",
			nil,
			nil,
		),
		pendingDesc: prometheus.NewDesc(
			"eth_pending_block_transactions",
			"the number of transactions in a pending block",
			nil,
			nil,
		),
	}
}

func (collector *EthBlockTransactionCount) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.earliestDesc
	ch <- collector.latestDesc
	ch <- collector.pendingDesc
}

func (collector *EthBlockTransactionCount) Collect(ch chan<- prometheus.Metric) {
	var wg sync.WaitGroup
	wg.Add(3)

	collect := func(tag string, desc *prometheus.Desc) {
		defer wg.Done()
		collector.collectByTag(tag, desc, ch)
	}

	go collect("earliest", collector.earliestDesc)
	go collect("latest", collector.latestDesc)
	go collect("pending", collector.pendingDesc)

	wg.Wait()
}
