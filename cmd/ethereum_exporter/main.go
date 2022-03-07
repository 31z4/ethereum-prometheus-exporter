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
	"strings"

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
	erc20ContractAddresses := flag.String("erc20ContractAddresses", "", "Comma-separated list of hexa ERC-20 contract addresses to listen for events.")
	startBlockNumber := flag.Uint64("startBlockNumber", 0, "block number from where to start watching events")
	addr := flag.String("addr", ":9368", "listen address")
	ver := flag.Bool("v", false, "print version number and exit")

	flag.Parse()
	if len(flag.Args()) > 0 {
		flag.Usage()
	}

	if *ver {
		fmt.Println(version)
		os.Exit(0)
	}

	var erc20Addresses []common.Address
	stringAddresses := strings.Split(*erc20ContractAddresses, ",")
	for _, stringAddr := range stringAddresses {
		erc20Addresses = append(erc20Addresses, common.HexToAddress(stringAddr))
	}
	log.Printf("Detected %d ERC-20 smart contract to monitor\n", len(erc20Addresses))

	rpc, err := rpc.Dial(*url)
	if err != nil {
		log.Fatalf("failed to create RPC client: %v", err)
	}

	client, err := ethclient.Dial(*url)
	if err != nil {
		log.Fatalf("failed to create ETH client: %v", err)
	}

	if startBlockNumber == nil || *startBlockNumber == 0 {
		log.Printf("Setting startBlockNumber to current block num")
		lastBlock, err := client.BlockNumber(context.Background())
		if err != nil {
			log.Fatalf("failed to get last block number: %v", err)
		}
		log.Printf("last block number: %d\n", lastBlock)
		*startBlockNumber = lastBlock
	}

	coll, err := collector.NewERC20TransferEvent(client, erc20Addresses, *startBlockNumber)
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
	)

	handler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{
		ErrorLog:      log.New(os.Stderr, log.Prefix(), log.Flags()),
		ErrorHandling: promhttp.ContinueOnError,
	})

	http.Handle("/metrics", handler)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
