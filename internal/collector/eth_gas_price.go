package collector

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/prometheus/client_golang/prometheus"
)

type EthGasPrice struct {
	rpc  *rpc.Client
	desc *prometheus.Desc
}

func NewEthGasPrice(rpc *rpc.Client) *EthGasPrice {
	return &EthGasPrice{
		rpc: rpc,
		desc: prometheus.NewDesc(
			"eth_gas_price",
			"current gas price in wei",
			nil,
			nil,
		),
	}
}

func (collector *EthGasPrice) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.desc
}

func (collector *EthGasPrice) Collect(ch chan<- prometheus.Metric) {
	var result hexutil.Big
	if err := collector.rpc.Call(&result, "eth_gasPrice"); err != nil {
		ch <- prometheus.NewInvalidMetric(collector.desc, err)
		return
	}

	i := (*big.Int)(&result)
	value, _ := new(big.Float).SetInt(i).Float64()
	ch <- prometheus.MustNewConstMetric(collector.desc, prometheus.GaugeValue, value)
}
