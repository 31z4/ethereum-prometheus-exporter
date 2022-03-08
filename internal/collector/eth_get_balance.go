package collector

import (
	"github.com/31z4/ethereum-prometheus-exporter/internal/config"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
)

type EthGetBalance struct {
	rpc     *rpc.Client
	address common.Address
	desc    *prometheus.Desc
}

func NewEthGetBalance(rpc *rpc.Client, wallet config.WalletTarget) *EthGetBalance {
	return &EthGetBalance{
		rpc:     rpc,
		address: common.HexToAddress(wallet.Addr),
		desc: prometheus.NewDesc(
			"eth_get_balance",
			"get balance",
			nil,
			map[string]string{"wallet_name": wallet.Name},
		),
	}
}

func (collector *EthGetBalance) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.desc
}

func (collector *EthGetBalance) Collect(ch chan<- prometheus.Metric) {
	var result hexutil.Uint64
	if err := collector.rpc.Call(&result, "eth_getBalance", collector.address, "latest"); err != nil {
		wErr := errors.Wrap(err, "failed to get Balance")
		ch <- prometheus.NewInvalidMetric(collector.desc, wErr)
		return
	}

	i := float64(result)
	ch <- prometheus.MustNewConstMetric(collector.desc, prometheus.GaugeValue, i)
}
