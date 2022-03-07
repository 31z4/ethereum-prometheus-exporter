package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"log"
	"net/http"
	"os"

	"github.com/31z4/ethereum-prometheus-exporter/internal/collector"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var version = "undefined"

func main() {
	flag.Usage = func() {
		const (
			usage = "Usage: ethereum_exporter [option] [arg]\n\n" +
				"Prometheus exporter for Ethereum client metrics\n\n" +
				"Options and arguments:\n"
		)

		fmt.Fprint(flag.CommandLine.Output(), usage)
		flag.PrintDefaults()

		os.Exit(2)
	}

	url := flag.String("url", "http://localhost:8545", "Ethereum JSON-RPC URL")
	erc20ContractAddress := flag.String("erc20.contractAddress", "", "ERC20 Contract Address to listen for events")
	startBlockNumber := flag.Uint64("startBlockNumber", 0, "block number from where to start watching events")
	addr := flag.String("addr", ":9368", "listen address")
	ver := flag.Bool("v", false, "print version number and exit")
	walletAddress := flag.String("address.checkBalance", "", "Wallet address to check balance")

	flag.Parse()
	if len(flag.Args()) > 0 {
		flag.Usage()
	}

	if *ver {
		fmt.Println(version)
		os.Exit(0)
	}

	erc20Address := common.HexToAddress(*erc20ContractAddress)

	rpc, err := rpc.Dial(*url)
	if err != nil {
		log.Fatalf("failed to create RPC client: %v", err)
	}

	collectorGetAddressBalance := collector.NewEthGetBalance(rpc, *walletAddress)

	client, err := ethclient.Dial(*url)
	if err != nil {
		log.Fatalf("failed to create ETH client: %v", err)
	}

	if startBlockNumber == nil {
		lastBlock, err := client.BlockNumber(context.Background())
		if err != nil {
			log.Fatalf("failed to get last block number: %v", err)
		}
		log.Printf("last block number: %d\n", lastBlock)
		*startBlockNumber = lastBlock
	}

	coll, err := collector.NewERC20TransferEvent(client, erc20Address, *startBlockNumber)
	if err != nil {
		log.Fatalf("failed to create erc20 transfer collector: %v", err)
	}

	registry := prometheus.NewPedanticRegistry()
	registry.MustRegister(
		collector.NewNetPeerCount(rpc),
		collector.NewEthBlockNumber(rpc),
		collector.NewEthBlockTimestamp(rpc),
		collector.NewEthGasPrice(rpc),
		collector.NewEthEarliestBlockTransactions(rpc),
		collector.NewEthLatestBlockTransactions(rpc),
		collector.NewEthPendingBlockTransactions(rpc),
		collector.NewEthHashrate(rpc),
		collector.NewEthSyncing(rpc),
		collector.NewParityNetPeers(rpc),
		coll,
		collectorGetAddressBalance,
	)

	handler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{
		ErrorLog:      log.New(os.Stderr, log.Prefix(), log.Flags()),
		ErrorHandling: promhttp.ContinueOnError,
	})

	http.Handle("/metrics", handler)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
