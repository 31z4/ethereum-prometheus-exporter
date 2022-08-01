package main

import (
	"flag"
	"fmt"
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
	addr := flag.String("addr", ":9368", "listen address")
	ver := flag.Bool("v", false, "print version number and exit")

	netpeercount := flag.Bool("netpeercount", true, "netpeercount metrics on/off switch")
	ethblocknumber := flag.Bool("ethblocknumber", true, "ethblocknumber metrics on/off switch")
	ethblocktimestamp := flag.Bool("ethblocktimestamp", false, "ethblocktimestamp metrics on/off switch")
	ethgasprice := flag.Bool("ethgasprice", true, "ethgasprice metrics on/off switch")
	ethearliestblocktransactions := flag.Bool("ethearliestblocktransactions", true, "ethearliestblocktransactions metrics on/off switch")
	ethlatestblocktransactions := flag.Bool("ethlatestblocktransactions", true, "ethlatestblocktransactions metrics on/off switch")
	ethpendingblocktransactions := flag.Bool("ethpendingblocktransactions", true, "ethpendingblocktransactions metrics on/off switch")
	ethhashrate := flag.Bool("ethhashrate", true, "ethhashrate metrics on/off switch")
	ethsyncing := flag.Bool("ethsyncing", false, "ethsyncing metrics on/off switch")
	paritynetpeers := flag.Bool("paritynetpeers", false, "paritynetpeers metrics on/off switch")

	flag.Parse()
	if len(flag.Args()) > 0 {
		flag.Usage()
	}

	if *ver {
		fmt.Println(version)
		os.Exit(0)
	}

	rpc, err := rpc.Dial(*url)
	if err != nil {
		log.Fatal(err)
	}

	registry := prometheus.NewPedanticRegistry()

	if *netpeercount {
		registry.MustRegister(
			collector.NewNetPeerCount(rpc),
		)
	}

	if *ethblocknumber {
		registry.MustRegister(
			collector.NewEthBlockNumber(rpc),
		)
	}

	if *ethblocktimestamp {
		registry.MustRegister(
			collector.NewEthBlockTimestamp(rpc),
		)
	}

	if *ethgasprice {
		registry.MustRegister(
			collector.NewEthGasPrice(rpc),
		)
	}

	if *ethearliestblocktransactions {
		registry.MustRegister(
			collector.NewEthEarliestBlockTransactions(rpc),
		)
	}

	if *ethlatestblocktransactions {
		registry.MustRegister(
			collector.NewEthLatestBlockTransactions(rpc),
		)
	}

	if *ethpendingblocktransactions {
		registry.MustRegister(
			collector.NewEthPendingBlockTransactions(rpc),
		)
	}

	if *ethhashrate {
		registry.MustRegister(
			collector.NewEthHashrate(rpc),
		)
	}

	if *ethsyncing {
		registry.MustRegister(
			collector.NewEthSyncing(rpc),
		)
	}

	if *paritynetpeers {
		registry.MustRegister(
			collector.NewParityNetPeers(rpc),
		)
	}

	//registry.MustRegister(
	//	collector.NewNetPeerCount(rpc),
	//	collector.NewEthBlockNumber(rpc),
	//	collector.NewEthBlockTimestamp(rpc),
	//	collector.NewEthGasPrice(rpc),
	//	collector.NewEthEarliestBlockTransactions(rpc),
	//	collector.NewEthLatestBlockTransactions(rpc),
	//	collector.NewEthPendingBlockTransactions(rpc),
	//	collector.NewEthHashrate(rpc),
	//	collector.NewEthSyncing(rpc),
	//	collector.NewParityNetPeers(rpc),
	//)

	handler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{
		ErrorLog:      log.New(os.Stderr, log.Prefix(), log.Flags()),
		ErrorHandling: promhttp.ContinueOnError,
	})

	http.Handle("/metrics", handler)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
