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
	contractClients  map[string]*token.TokenFilterer
	desc             *prometheus.Desc
	collectMutex     sync.Mutex
	lastQueriedBlock uint64
	bnGetter         BlockNumberGetter
}

func NewERC20TransferEvent(client PagedContractFilterer, contractAddresses []common.Address, nowBlockNumber uint64) (*ERC20TransferEvent, error) {
	clients := map[string]*token.TokenFilterer{}
	for _, contractAddress := range contractAddresses {
		filterer, err := token.NewTokenFilterer(contractAddress, client)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create ERC20 transfer evt collector")
		}
		clients[contractAddress.Hex()] = filterer
	}

	return &ERC20TransferEvent{
		contractClients: clients,
		desc: prometheus.NewDesc(
			"erc20_transfer_event",
			"ERC20 Transfer events count",
			[]string{"contract"},
			nil,
		),
		lastQueriedBlock: nowBlockNumber,
		bnGetter:         client,
	}, nil
}

func (col *ERC20TransferEvent) Describe(ch chan<- *prometheus.Desc) {
	ch <- col.desc
}

func (col *ERC20TransferEvent) doCollect(ch chan<- prometheus.Metric, currentBlockNumber uint64, contract string, client *token.TokenFilterer) {
	it, err := client.FilterTransfer(&bind.FilterOpts{
		Context: context.Background(),
		Start:   col.lastQueriedBlock,
		End:     &currentBlockNumber,
	}, nil, nil)
	if err != nil {
		wErr := errors.Wrapf(err, "failed to create transfer iterator for contract=[%s]", contract)
		ch <- prometheus.NewInvalidMetric(col.desc, wErr)
		return
	}

	// histogram summary to collect
	var count uint64 = 0
	var sum float64 = 0

	for {
		eventsLeft := it.Next()
		if !eventsLeft && it.Error() == nil {
			// Finished reading events, advance lastQueriedBlock and publish histogram data
			ch <- prometheus.MustNewConstHistogram(col.desc, count, sum, nil, contract)
			return
		} else if !eventsLeft {
			wErr := errors.Wrapf(err, "failed to read transfer event for contract=[%s]", contract)
			ch <- prometheus.NewInvalidMetric(col.desc, wErr)
			return
		}
		te := it.Event

		value, _ := new(big.Float).SetInt(te.Tokens).Float64()
		count += 1
		sum += value
	}

}

func (col *ERC20TransferEvent) Collect(ch chan<- prometheus.Metric) {
	col.collectMutex.Lock()
	defer col.collectMutex.Unlock()

	currentBlockNumber, err := col.bnGetter.BlockNumber(context.Background())
	if err != nil {
		wErr := errors.Wrap(err, "failed to get current block number")
		ch <- prometheus.NewInvalidMetric(col.desc, wErr)
		return
	}
	// INV: currentBlockNum >= lastQueriedblock

	wg := sync.WaitGroup{}
	for contract, client := range col.contractClients {
		wg.Add(1)

		go func() {
			defer wg.Done()
			col.doCollect(ch, currentBlockNumber, contract, client)
		}()
	}

	wg.Wait()

	// Improve error model, this will advance last seen block even if some client fails
	col.lastQueriedBlock = currentBlockNumber
}
