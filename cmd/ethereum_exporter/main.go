package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/31z4/ethereum-prometheus-exporter/internal/collector"
	"github.com/31z4/ethereum-prometheus-exporter/internal/config"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/prometheus/client_golang/prometheus"
	"log"
	"net/http"
	"os"

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

	configFile := flag.String("config", "", "path to config file")
	ver := flag.Bool("v", false, "print version number and exit")

	flag.Parse()
	if len(flag.Args()) > 0 {
		flag.Usage()
	}

	if *ver {
		fmt.Println(version)
		os.Exit(0)
	}

	cfg, err := config.ParseConfigFromFile(*configFile)
	if err != nil {
		log.Fatalf("Failed to read config file (%v): %v", configFile, err)
	}

	// Initiate clients
	rpc, err := rpc.Dial(cfg.General.EthProviderURL)
	if err != nil {
		log.Fatalf("failed to create RPC client: %v", err)
	}

	client, err := ethclient.Dial(cfg.General.EthProviderURL)
	if err != nil {
		log.Fatalf("failed to create ETH client: %v", err)
	}

	if cfg.General.StartBlockNumber == 0 {
		log.Printf("Setting startBlockNumber to current block num")
		lastBlock, err := client.BlockNumber(context.Background())
		if err != nil {
			log.Fatalf("failed to get last block number: %v", err)
		}
		log.Printf("last block number: %d\n", lastBlock)
		cfg.General.StartBlockNumber = lastBlock
	}

	// ERC-20 Targets
	var addresses []common.Address
	for _, target := range cfg.Target.ERC20 {
		addresses = append(addresses, common.HexToAddress(target.ContractAddr))
	}
	log.Printf("Detected %d ERC-20 smart contract(s) to monitor\n", len(addresses))

	coll, err := collector.NewERC20TransferEvent(client, addresses, cfg.General.StartBlockNumber)
	if err != nil {
		log.Fatalf("failed to create erc20 transfer collector: %v", err)
	}

	// Wallet  Target
	collectorGetAddressBalance := collector.NewEthGetBalance(rpc, cfg.Target.Wallet.Addr)

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
	log.Fatal(http.ListenAndServe(cfg.General.ServerURL, nil))
}
