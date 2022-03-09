package eth

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/ethereum/go-ethereum/rpc"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

func TestEthBlockTimestampCollectError(t *testing.T) {
	rpc, err := rpc.DialHTTP("http://localhost")
	if err != nil {
		t.Fatalf("rpc connection error: %#v", err)
	}

	collector := NewEthBlockTimestamp(rpc)
	ch := make(chan prometheus.Metric, 1)

	collector.Collect(ch)
	close(ch)

	if got := len(ch); got != 1 {
		t.Fatalf("got %v, want 1", got)
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

func TestEthBlockTimestampCollect(t *testing.T) {
	rpcServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(`{"result": {"number": "0xaca4b5", "timestamp": "0x5fbba343"}}`))
		if err != nil {
			t.Fatalf("could not write a response: %#v", err)
		}
	}))

	defer rpcServer.Close()

	rpc, err := rpc.DialHTTP(rpcServer.URL)
	if err != nil {
		t.Fatalf("rpc connection error: %#v", err)
	}

	collector := NewEthBlockTimestamp(rpc)
	ch := make(chan prometheus.Metric, 1)

	collector.Collect(ch)
	close(ch)

	if got := len(ch); got != 1 {
		t.Fatalf("got %v, want 1", got)
	}

	var metric dto.Metric
	for result := range ch {
		if err := result.Write(&metric); err != nil {
			t.Fatalf("expected metric, got %#v", err)
		}
		if got := len(metric.Label); got > 0 {
			t.Fatalf("expected 0 labels, got %d", got)
		}
		if got := *metric.Gauge.Value; got != 1606132547 {
			t.Fatalf("got %v, want 1606132547 ", got)
		}
	}
}
