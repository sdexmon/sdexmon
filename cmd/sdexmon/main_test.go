package main

import (
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stellar/go/clients/horizonclient"
	"github.com/stellar/go/txnbuild"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return fn(request)
}

func TestFetchOrderbookUsesOneCanonicalRequest(t *testing.T) {
	var requests atomic.Int32
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		requests.Add(1)
		if r.URL.Path != "/order_book" {
			t.Errorf("unexpected request path %q", r.URL.Path)
		}
		if got := r.URL.Query().Get("selling_asset_type"); got != "native" {
			t.Errorf("selling_asset_type = %q, want native", got)
		}
		body := `{
			"bids":[{"price_r":{"n":1,"d":1},"price":"1.0000000","amount":"2.0000000"}],
			"asks":[{"price_r":{"n":2,"d":1},"price":"2.0000000","amount":"3.0000000"}],
			"base":{"asset_type":"native"},
			"counter":{"asset_type":"credit_alphanum4","asset_code":"USDC","asset_issuer":"GA5ZSEJYB37JRC5AVCIA5MOP4RHTM335X2KGX3IHOJAPP5RE34K4KZVN"}
		}`
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(body)),
			Request:    r,
		}, nil
	})

	client := &horizonclient.Client{
		HorizonURL: "https://horizon.example",
		HTTP:       &http.Client{Transport: transport},
	}
	quote := txnbuild.CreditAsset{
		Code:   "USDC",
		Issuer: "GA5ZSEJYB37JRC5AVCIA5MOP4RHTM335X2KGX3IHOJAPP5RE34K4KZVN",
	}
	message := fetchOrderbookCmd(client, txnbuild.NativeAsset{}, quote)()
	data, ok := message.(orderbookDataMsg)
	if !ok {
		t.Fatalf("message type = %T, want orderbookDataMsg", message)
	}
	if requests.Load() != 1 {
		t.Fatalf("requests = %d, want 1", requests.Load())
	}
	if len(data.ob.Bids) != 1 || len(data.ob.Asks) != 1 {
		t.Fatalf("book sides = %d bids/%d asks, want 1/1", len(data.ob.Bids), len(data.ob.Asks))
	}
}

func TestAssetsForPairPreservesIssuer(t *testing.T) {
	base := txnbuild.CreditAsset{Code: "USD", Issuer: "GAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAWHF"}
	quote := txnbuild.NativeAsset{}
	pair := pairOption{Base: "USD", Quote: "XLM", BaseAsset: base, QuoteAsset: quote}

	gotBase, gotQuote, ok := assetsForPair(pair)
	if !ok {
		t.Fatal("assetsForPair returned invalid pair")
	}
	if assetString(gotBase) != assetString(base) || assetString(gotQuote) != "native" {
		t.Fatalf("assetsForPair lost asset identity: %q / %q", assetString(gotBase), assetString(gotQuote))
	}
}
