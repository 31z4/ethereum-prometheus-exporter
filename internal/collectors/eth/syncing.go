package eth

import (
	"encoding/json"
	"errors"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/prometheus/client_golang/prometheus"
)

type EthSyncing struct {
	rpc          *rpc.Client
	startingDesc *prometheus.Desc
	currentDesc  *prometheus.Desc
	highestDesc  *prometheus.Desc
}

type syncingResult struct {
	StartingBlock hexutil.Uint64
	CurrentBlock  hexutil.Uint64
	HighestBlock  hexutil.Uint64
}

func NewEthSyncing(rpc *rpc.Client, label string) *EthSyncing {
	return &EthSyncing{
		rpc: rpc,
		startingDesc: prometheus.NewDesc(
			"eth_sync_starting",
			"block number at which current import started",
			nil,
			map[string]string{"blockchain_name": label},
		),
		currentDesc: prometheus.NewDesc(
			"eth_sync_current",
			"number of most recent block",
			nil,
			map[string]string{"blockchain_name": label},
		),
		highestDesc: prometheus.NewDesc(
			"eth_sync_highest",
			"estimated number of highest block",
			nil,
			map[string]string{"blockchain_name": label},
		),
	}
}

func (collector *EthSyncing) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.startingDesc
	ch <- collector.currentDesc
	ch <- collector.highestDesc
}

func (collector *EthSyncing) Collect(ch chan<- prometheus.Metric) {
	var raw json.RawMessage
	if err := collector.rpc.Call(&raw, "eth_syncing"); err != nil {
		ch <- prometheus.NewInvalidMetric(collector.startingDesc, err)
		ch <- prometheus.NewInvalidMetric(collector.currentDesc, err)
		ch <- prometheus.NewInvalidMetric(collector.highestDesc, err)
		return
	}

	var syncing bool
	if err := json.Unmarshal(raw, &syncing); err == nil {
		err = errors.New("not syncing")
		ch <- prometheus.NewInvalidMetric(collector.startingDesc, err)
		ch <- prometheus.NewInvalidMetric(collector.currentDesc, err)
		ch <- prometheus.NewInvalidMetric(collector.highestDesc, err)
		return
	}

	var result *syncingResult
	if err := json.Unmarshal(raw, &result); err != nil {
		ch <- prometheus.NewInvalidMetric(collector.startingDesc, err)
		ch <- prometheus.NewInvalidMetric(collector.currentDesc, err)
		ch <- prometheus.NewInvalidMetric(collector.highestDesc, err)
		return
	}

	value := float64(result.StartingBlock)
	ch <- prometheus.MustNewConstMetric(collector.startingDesc, prometheus.GaugeValue, value)
	value = float64(result.CurrentBlock)
	ch <- prometheus.MustNewConstMetric(collector.currentDesc, prometheus.GaugeValue, value)
	value = float64(result.HighestBlock)
	ch <- prometheus.MustNewConstMetric(collector.highestDesc, prometheus.GaugeValue, value)
}
