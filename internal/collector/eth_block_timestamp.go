package collector

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/prometheus/client_golang/prometheus"
)

type EthBlockTimestamp struct {
	rpc  *rpc.Client
	desc *prometheus.Desc
}

type blockResult struct {
	Timestamp hexutil.Uint64
}

func NewEthBlockTimestamp(rpc *rpc.Client, label string) *EthBlockTimestamp {
	return &EthBlockTimestamp{
		rpc: rpc,
		desc: prometheus.NewDesc(
			"eth_block_timestamp",
			"timestamp of the most recent block",
			nil,
			map[string]string{"blockchain_name": label},
		),
	}
}

func (collector *EthBlockTimestamp) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.desc
}

func (collector *EthBlockTimestamp) Collect(ch chan<- prometheus.Metric) {
	var result *blockResult

	if err := collector.rpc.Call(&result, "eth_getBlockByNumber", "latest", false); err != nil {
		ch <- prometheus.NewInvalidMetric(collector.desc, err)
		return
	}

	value := float64(result.Timestamp)

	ch <- prometheus.MustNewConstMetric(collector.desc, prometheus.GaugeValue, value)
}
