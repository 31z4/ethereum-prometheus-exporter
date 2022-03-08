package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/31z4/ethereum-prometheus-exporter/internal/collector"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var version = "undefined"

type ERC20Target struct {
	Name         string `yaml:"name"`
	ContractAddr string `yaml:"contract"`
}

type WalletTarget struct {
	Addr string `yaml:"address"`
}

type Config struct {
	General struct {
		EthProviderURL   string `yaml:"eth_provider_url"`
		ServerURL        string `yaml:"server_url"`
		StartBlockNumber uint64 `yaml:"start_block_number"`
	} `yaml:"general"`
	Target struct {
		ERC20  []ERC20Target `yaml:"erc20"`
		Wallet WalletTarget  `yaml:"wallet"`
	} `yaml:"target"`
}

func parseConfigFromFile(path string) (*Config, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	config := new(Config)
	err = yaml.Unmarshal(bytes, config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

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

	config, err := parseConfigFromFile(*configFile)
	if err != nil {
		log.Fatalf("Failed to read config file (%v): %v", configFile, err)
	}

	// Initiate clients
	rpc, err := rpc.Dial(config.General.EthProviderURL)
	if err != nil {
		log.Fatalf("failed to create RPC client: %v", err)
	}

	client, err := ethclient.Dial(config.General.EthProviderURL)
	if err != nil {
		log.Fatalf("failed to create ETH client: %v", err)
	}

	if config.General.StartBlockNumber == 0 {
		log.Printf("Setting startBlockNumber to current block num")
		lastBlock, err := client.BlockNumber(context.Background())
		if err != nil {
			log.Fatalf("failed to get last block number: %v", err)
		}
		log.Printf("last block number: %d\n", lastBlock)
		config.General.StartBlockNumber = lastBlock
	}

	// ERC-20 Targets
	var addresses []common.Address
	for _, target := range config.Target.ERC20 {
		addresses = append(addresses, common.HexToAddress(target.ContractAddr))
	}
	log.Printf("Detected %d ERC-20 smart contract(s) to monitor\n", len(addresses))

	coll, err := collector.NewERC20TransferEvent(client, addresses, config.General.StartBlockNumber)
	if err != nil {
		log.Fatalf("failed to create erc20 transfer collector: %v", err)
	}

	// Wallet  Target
	collectorGetAddressBalance := collector.NewEthGetBalance(rpc, config.Target.Wallet.Addr)

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
	log.Fatal(http.ListenAndServe(config.General.ServerURL, nil))
}
