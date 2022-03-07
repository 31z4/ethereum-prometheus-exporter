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

type BlockNumberGetter interface {
	BlockNumber(ctx context.Context) (uint64, error)
}

type PagedContractFilterer interface {
	BlockNumberGetter
	bind.ContractFilterer
}

type ERC20TransferEvent struct {
	client           *token.TokenFilterer
	desc             *prometheus.Desc
	collectMutex     sync.Mutex
	lastQueriedBlock uint64
	bnGetter         BlockNumberGetter
}

func NewERC20TransferEvent(client PagedContractFilterer, contractAddress common.Address, nowBlockNumber uint64) (*ERC20TransferEvent, error) {
	filterer, err := token.NewTokenFilterer(contractAddress, client)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create ERC20 transfer evt collector")
	}

	return &ERC20TransferEvent{
		client: filterer,
		desc: prometheus.NewDesc(
			"erc20_transfer_event",
			"ERC20 Transfer events count",
			[]string{"from", "to"},
			prometheus.Labels{"contract": contractAddress.Hex()},
		),
		lastQueriedBlock: nowBlockNumber,
		bnGetter:         client,
	}, nil
}

func (col *ERC20TransferEvent) Describe(ch chan<- *prometheus.Desc) {
	ch <- col.desc
}

func (col *ERC20TransferEvent) Collect(ch chan<- prometheus.Metric) {
	col.collectMutex.Lock()
	defer col.collectMutex.Unlock()

	currentBlockNum, err := col.bnGetter.BlockNumber(context.Background())
	if err != nil {
		wErr := errors.Wrap(err, "failed to get current block number")
		ch <- prometheus.NewInvalidMetric(col.desc, wErr)
		return
	}

	// INV: currentBlockNum >= lastQueriedblock

	it, err := col.client.FilterTransfer(&bind.FilterOpts{
		Context: context.Background(),
		Start:   col.lastQueriedBlock,
		End:     &currentBlockNum,
	}, nil, nil)
	if err != nil {
		wErr := errors.Wrap(err, "failed to create transfer iterator")
		ch <- prometheus.NewInvalidMetric(col.desc, wErr)
		return
	}

	for {
		eventsLeft := it.Next()
		if !eventsLeft && it.Error() == nil {
			// Finished reading events, advance lastQueriedBlock
			col.lastQueriedBlock = currentBlockNum
			return
		} else if !eventsLeft {
			wErr := errors.Wrap(err, "failed to read transfer event")
			ch <- prometheus.NewInvalidMetric(col.desc, wErr)
			return
		}
		te := it.Event
		value, _ := new(big.Float).SetInt(te.Tokens).Float64()
		ch <- prometheus.MustNewConstHistogram(col.desc, prometheus.GaugeValue, value, te.From.Hex(), te.To.Hex())
		// FIXME: Maybe I should read all events, and then advance lastQueriedBlock, to avoid re-reading events
	}
}
