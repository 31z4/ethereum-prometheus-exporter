package collector

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/ethereum/go-ethereum/rpc"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

func TestEthSyncingCollectError(t *testing.T) {
	rpc, err := rpc.DialHTTP("http://localhost")
	if err != nil {
		t.Fatalf("rpc connection error: %#v", err)
	}

	collector := NewEthSyncing(rpc, blockchainName)
	ch := make(chan prometheus.Metric, 3)

	collector.Collect(ch)
	close(ch)

	if got := len(ch); got != 3 {
		t.Fatalf("got %v, want 3", got)
	}

	var metric dto.Metric
	for result := range ch {
		err := result.Write(&metric)
		if err == nil {
			t.Fatalf("expected invalid metric, got %#v", metric)
		}
		if _, ok := err.(*url.Error); !ok {
			t.Fatalf("unexpected error %#v", err)
		}
	}
}

func TestEthSyncingCollectNotSyncing(t *testing.T) {
	rpcServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(`{"result": false}"`))
		if err != nil {
			t.Fatalf("could not write a response: %#v", err)
		}
	}))
	defer rpcServer.Close()

	rpc, err := rpc.DialHTTP(rpcServer.URL)
	if err != nil {
		t.Fatalf("rpc connection error: %#v", err)
	}

	collector := NewEthSyncing(rpc, blockchainName)
	ch := make(chan prometheus.Metric, 3)

	collector.Collect(ch)
	close(ch)

	if got := len(ch); got != 3 {
		t.Fatalf("got %v, want 3", got)
	}

	var metric dto.Metric
	for result := range ch {
		err := result.Write(&metric)
		if err == nil {
			t.Fatalf("expected invalid metric, got %#v", metric)
		}
		if err.Error() != "not syncing" {
			t.Fatalf("unexpected error %#v", err)
		}
	}
}

func TestEthSyncingCollectUnmarshalError(t *testing.T) {
	rpcServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(`{"result": "test"}"`))
		if err != nil {
			t.Fatalf("could not write a response: %#v", err)
		}
	}))
	defer rpcServer.Close()

	rpc, err := rpc.DialHTTP(rpcServer.URL)
	if err != nil {
		t.Fatalf("rpc connection error: %#v", err)
	}

	collector := NewEthSyncing(rpc, blockchainName)
	ch := make(chan prometheus.Metric, 3)

	collector.Collect(ch)
	close(ch)

	if got := len(ch); got != 3 {
		t.Fatalf("got %v, want 3", got)
	}

	var metric dto.Metric
	for result := range ch {
		err := result.Write(&metric)
		if err == nil {
			t.Fatalf("expected invalid metric, got %#v", metric)
		}
		if _, ok := err.(*json.UnmarshalTypeError); !ok {
			t.Fatalf("unexpected error %#v", err)
		}
	}
}

func TestEthSyncingCollect(t *testing.T) {
	rpcServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(`{"result": {"startingBlock": "0x384", "currentBlock": "0x386", "highestBlock": "0x454"}}"`))
		if err != nil {
			t.Fatalf("could not write a response: %#v", err)
		}
	}))
	defer rpcServer.Close()

	rpc, err := rpc.DialHTTP(rpcServer.URL)
	if err != nil {
		t.Fatalf("rpc connection error: %#v", err)
	}

	collector := NewEthSyncing(rpc, blockchainName)
	ch := make(chan prometheus.Metric, 3)

	collector.Collect(ch)
	close(ch)

	if got := len(ch); got != 3 {
		t.Fatalf("got %v, want 3", got)
	}

	var (
		metric dto.Metric
		result prometheus.Metric
	)

	result = <-ch
	if err := result.Write(&metric); err != nil {
		t.Fatalf("expected metric, got %#v", err)
	}
	if got := len(metric.Label); got != 1 {
		t.Fatalf("expected 1 label, got %d", got)
	}
	if got := *metric.Gauge.Value; got != 900 {
		t.Fatalf("got %v, want 900", got)
	}

	result = <-ch
	if err := result.Write(&metric); err != nil {
		t.Fatalf("expected metric, got %#v", err)
	}
	if got := len(metric.Label); got != 1 {
		t.Fatalf("expected 1 label, got %d", got)
	}
	if got := *metric.Gauge.Value; got != 902 {
		t.Fatalf("got %v, want 902", got)
	}

	result = <-ch
	if err := result.Write(&metric); err != nil {
		t.Fatalf("expected metric, got %#v", err)
	}
	if got := len(metric.Label); got != 1 {
		t.Fatalf("expected 1 label, got %d", got)
	}
	if got := *metric.Gauge.Value; got != 1108 {
		t.Fatalf("got %v, want 1108", got)
	}
}
