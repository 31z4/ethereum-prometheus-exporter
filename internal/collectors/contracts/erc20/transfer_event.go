package erc20

import (
	"context"
	"github.com/thepalbi/ethereum-prometheus-exporter/internal/config"
	"log"
	"math"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/thepalbi/ethereum-prometheus-exporter/token"
)

type BlockNumberGetter interface {
	BlockNumber(ctx context.Context) (uint64, error)
}

type ContractClient interface {
	BlockNumberGetter
	bind.ContractFilterer
	bind.ContractCaller
}

type contractInfo struct {
	Address  string
	Symbol   string
	Decimals uint8
	Label    string
}

type ERC20TransferEvent struct {
	contractClients  map[*contractInfo]*token.TokenFilterer
	desc             *prometheus.Desc
	collectMutex     sync.Mutex
	lastQueriedBlock uint64
	bnGetter         BlockNumberGetter
}

func getContractInfo(contractAddr common.Address, contractClient bind.ContractCaller, label string) (*contractInfo, error) {
	contractCaller, err := token.NewTokenCaller(contractAddr, contractClient)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get contract info for %s", contractAddr.Hex())
	}
	symbol, err := contractCaller.Symbol(nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get symbol for %s", contractAddr.Hex())
	}
	decimals, err := contractCaller.Decimals(nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get decimals for %s", contractAddr.Hex())
	}
	return &contractInfo{
		Address:  contractAddr.Hex(),
		Symbol:   symbol,
		Decimals: decimals,
		Label:    label,
	}, nil
}

func NewERC20TransferEvent(client ContractClient, contractAddresses []config.ERC20Target, nowBlockNumber uint64) (*ERC20TransferEvent, error) {
	clients := map[*contractInfo]*token.TokenFilterer{}
	for _, contractAddress := range contractAddresses {
		address := common.HexToAddress(contractAddress.ContractAddr)
		filterer, err := token.NewTokenFilterer(address, client)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create ERC20 transfer evt collector")
		}
		info, err := getContractInfo(address, client, contractAddress.Name)
		if err != nil {
			return nil, err
		}

		log.Printf("Got info for %s, symbol %s, address_name %s\n", info.Address, info.Symbol, info.Label)
		clients[info] = filterer
	}

	return &ERC20TransferEvent{
		contractClients: clients,
		desc: prometheus.NewDesc(
			"erc20_transfer_event",
			"ERC20 Transfer events count",
			[]string{"contract", "symbol", "address_name"},
			nil,
		),
		lastQueriedBlock: nowBlockNumber,
		bnGetter:         client,
	}, nil
}

func (col *ERC20TransferEvent) Describe(ch chan<- *prometheus.Desc) {
	ch <- col.desc
}

func (col *ERC20TransferEvent) doCollect(ch chan<- prometheus.Metric, currentBlockNumber uint64, info *contractInfo, client *token.TokenFilterer) {
	it, err := client.FilterTransfer(&bind.FilterOpts{
		Context: context.Background(),
		Start:   col.lastQueriedBlock,
		End:     &currentBlockNumber,
	}, nil, nil)
	if err != nil {
		wErr := errors.Wrapf(err, "failed to create transfer iterator for contract=[%s]", info.Address)
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
			ch <- prometheus.MustNewConstHistogram(col.desc, count, sum, nil, info.Address, info.Symbol, info.Label)
			return
		} else if !eventsLeft {
			wErr := errors.Wrapf(err, "failed to read transfer event for contract=[%s]", info.Address)
			ch <- prometheus.NewInvalidMetric(col.desc, wErr)
			return
		}
		te := it.Event

		value, _ := new(big.Float).SetInt(te.Tokens).Float64()
		count += 1
		sum += value / math.Pow10(int(info.Decimals))
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
	for contrInfo, client := range col.contractClients {
		wg.Add(1)

		go func(contrInfo *contractInfo, client *token.TokenFilterer) {
			defer wg.Done()
			col.doCollect(ch, currentBlockNumber, contrInfo, client)
		}(contrInfo, client)
	}

	wg.Wait()

	// Improve error model, this will advance last seen block even if some client fails
	col.lastQueriedBlock = currentBlockNumber
}
