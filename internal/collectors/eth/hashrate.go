package eth

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/prometheus/client_golang/prometheus"
)

type EthHashrate struct {
	rpc  *rpc.Client
	desc *prometheus.Desc
}

func NewEthHashrate(rpc *rpc.Client, label string) *EthHashrate {
	return &EthHashrate{
		rpc: rpc,
		desc: prometheus.NewDesc(
			"eth_hashrate",
			"hashes per second that this node is mining with",
			nil,
			map[string]string{"blockchain_name": label},
		),
	}
}

func (collector *EthHashrate) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.desc
}

func (collector *EthHashrate) Collect(ch chan<- prometheus.Metric) {
	var result hexutil.Uint64
	if err := collector.rpc.Call(&result, "eth_hashrate"); err != nil {
		ch <- prometheus.NewInvalidMetric(collector.desc, err)
		return
	}

	value := float64(result)
	ch <- prometheus.MustNewConstMetric(collector.desc, prometheus.GaugeValue, value)
}
