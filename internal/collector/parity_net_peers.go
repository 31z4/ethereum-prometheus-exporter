package collector

import (
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
)

type ParityNetPeers struct {
	rpc           *rpc.Client
	activeDesc    *prometheus.Desc
	connectedDesc *prometheus.Desc
}

type peersResult struct {
	Active    uint64
	Connected uint64
}

func NewParityNetPeers(rpc *rpc.Client, label string) *ParityNetPeers {
	return &ParityNetPeers{
		rpc: rpc,
		activeDesc: prometheus.NewDesc(
			"parity_net_active_peers",
			"number of active peers",
			nil,
			map[string]string{"blockchain_name": label},
		),
		connectedDesc: prometheus.NewDesc(
			"parity_net_connected_peers",
			"number of peers currently connected to this client",
			nil,
			map[string]string{"blockchain_name": label},
		),
	}
}

func (collector *ParityNetPeers) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.activeDesc
	ch <- collector.connectedDesc
}

func (collector *ParityNetPeers) Collect(ch chan<- prometheus.Metric) {
	var result *peersResult
	if err := collector.rpc.Call(&result, "parity_netPeers"); err != nil {
		wErr := errors.Wrap(err, "parity metrics are only available in OpenEthereum")
		ch <- prometheus.NewInvalidMetric(collector.activeDesc, wErr)
		ch <- prometheus.NewInvalidMetric(collector.connectedDesc, wErr)
		return
	}

	value := float64(result.Active)
	ch <- prometheus.MustNewConstMetric(collector.activeDesc, prometheus.GaugeValue, value)
	value = float64(result.Connected)
	ch <- prometheus.MustNewConstMetric(collector.connectedDesc, prometheus.GaugeValue, value)
}
