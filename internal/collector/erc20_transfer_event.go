package collector

import (
	"context"
	"github.com/31z4/ethereum-prometheus-exporter/token"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"math/big"
	"sync"
)

type ERC20TransferEvent struct {
	filterer         *token.TokenFilterer
	desc             *prometheus.Desc
	collectMutex     sync.Mutex
	lastQueriedBlock uint64
}

func NewERC20TransferEvent(client bind.ContractFilterer, contractAddress common.Address, nowBlockNumber uint64) (*ERC20TransferEvent, error) {
	filterer, err := token.NewTokenFilterer(contractAddress, client)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create ERC20 transfer evt collector")
	}

	return &ERC20TransferEvent{
		filterer: filterer,
		desc: prometheus.NewDesc(
			"erc20_transfer_event",
			"ERC20 Transfer events count",
			[]string{"from", "to"},
			prometheus.Labels{"contract": contractAddress.Hex()},
		),
		lastQueriedBlock: nowBlockNumber,
	}, nil
}

func (col *ERC20TransferEvent) Describe(ch chan<- *prometheus.Desc) {
	ch <- col.desc
}

func (col *ERC20TransferEvent) Collect(ch chan<- prometheus.Metric) {
	col.collectMutex.Lock()
	defer col.collectMutex.Unlock()

	it, err := col.filterer.FilterTransfer(&bind.FilterOpts{
		Context: context.Background(),
		Start:   col.lastQueriedBlock,
	}, nil, nil)
	if err != nil {
		wErr := errors.Wrap(err, "failed to create transfer iterator")
		ch <- prometheus.NewInvalidMetric(col.desc, wErr)
		return
	}

	for {
		ok := it.Next()
		if !ok && it.Error() == nil {
			// Finished reading events
			return
		} else if !ok {
			wErr := errors.Wrap(err, "failed to read transfer event")
			ch <- prometheus.NewInvalidMetric(col.desc, wErr)
			return
		}
		te := it.Event
		value, _ := new(big.Float).SetInt(te.Tokens).Float64()
		ch <- prometheus.MustNewConstMetric(col.desc, prometheus.CounterValue, value, te.From.Hex(), te.To.Hex())
	}
}
