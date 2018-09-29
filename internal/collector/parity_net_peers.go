package collector

import (
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/prometheus/client_golang/prometheus"
)

type ParityNetPeers struct {
	rpc  *rpc.Client
	desc *prometheus.Desc
}

type peersResult struct {
	Active    uint64
	Connected uint64
}

func NewParityNetPeers(rpc *rpc.Client) *ParityNetPeers {
	return &ParityNetPeers{
		rpc: rpc,
		desc: prometheus.NewDesc(
			"parity_net_peers",
			"the number of peers currently connected to the client",
			[]string{"status"},
			nil,
		),
	}
}

func (collector *ParityNetPeers) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.desc
}

func (collector *ParityNetPeers) Collect(ch chan<- prometheus.Metric) {
	var result *peersResult
	if err := collector.rpc.Call(&result, "parity_netPeers"); err != nil {
		ch <- prometheus.NewInvalidMetric(collector.desc, err)
		return
	}

	value := float64(result.Active)
	ch <- prometheus.MustNewConstMetric(collector.desc, prometheus.GaugeValue, value, "active")
	value = float64(result.Connected)
	ch <- prometheus.MustNewConstMetric(collector.desc, prometheus.GaugeValue, value, "connected")
}
