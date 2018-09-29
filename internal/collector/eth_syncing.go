package collector

import (
	"encoding/json"
	"errors"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/prometheus/client_golang/prometheus"
)

type EthSyncing struct {
	rpc  *rpc.Client
	desc *prometheus.Desc
}

type syncingResult struct {
	StartingBlock hexutil.Uint64
	CurrentBlock  hexutil.Uint64
	HighestBlock  hexutil.Uint64
}

func NewEthSyncing(rpc *rpc.Client) *EthSyncing {
	return &EthSyncing{
		rpc: rpc,
		desc: prometheus.NewDesc(
			"eth_syncing",
			"data about the sync status",
			[]string{"block"},
			nil,
		),
	}
}

func (collector *EthSyncing) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.desc
}

func (collector *EthSyncing) Collect(ch chan<- prometheus.Metric) {
	var raw json.RawMessage
	if err := collector.rpc.Call(&raw, "eth_syncing"); err != nil {
		ch <- prometheus.NewInvalidMetric(collector.desc, err)
		return
	}

	var syncing bool
	if err := json.Unmarshal(raw, &syncing); err == nil {
		ch <- prometheus.NewInvalidMetric(collector.desc, errors.New("not syncing"))
		return
	}

	var result *syncingResult
	if err := json.Unmarshal(raw, &result); err != nil {
		ch <- prometheus.NewInvalidMetric(collector.desc, err)
		return
	}

	value := float64(result.CurrentBlock)
	ch <- prometheus.MustNewConstMetric(collector.desc, prometheus.GaugeValue, value, "current")
	value = float64(result.HighestBlock)
	ch <- prometheus.MustNewConstMetric(collector.desc, prometheus.GaugeValue, value, "highest")
	value = float64(result.StartingBlock)
	ch <- prometheus.MustNewConstMetric(collector.desc, prometheus.GaugeValue, value, "starting")
}
