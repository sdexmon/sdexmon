package stellar

import (
	"fmt"

	"github.com/sdexmon/sdexmon/internal/models"
	"github.com/stellar/go/clients/horizonclient"
	"github.com/stellar/go/txnbuild"
)

// FetchPairConfirmationData fetches market data for pair confirmation screen
func FetchPairConfirmationData(client *horizonclient.Client, assetA, assetB txnbuild.Asset, poolIDsMap map[string]string) (*models.PairConfirmationData, error) {
	data := &models.PairConfirmationData{
		AssetA:    assetA,
		AssetB:    assetB,
		BestBid:   "N/A",
		BestAsk:   "N/A",
		LPLockedA: "N/A",
		LPLockedB: "N/A",
		LPPoolID:  "",
	}

	// Fetch order book
	obReq := horizonclient.OrderBookRequest{}
	
	// Apply selling asset
	if assetA.IsNative() {
		obReq.SellingAssetType = horizonclient.AssetTypeNative
	} else {
		obReq.SellingAssetType = horizonclient.AssetType4
		obReq.SellingAssetCode = assetA.GetCode()
		obReq.SellingAssetIssuer = assetA.GetIssuer()
		if len(assetA.GetCode()) > 4 {
			obReq.SellingAssetType = horizonclient.AssetType12
		}
	}

	// Apply buying asset
	if assetB.IsNative() {
		obReq.BuyingAssetType = horizonclient.AssetTypeNative
	} else {
		obReq.BuyingAssetType = horizonclient.AssetType4
		obReq.BuyingAssetCode = assetB.GetCode()
		obReq.BuyingAssetIssuer = assetB.GetIssuer()
		if len(assetB.GetCode()) > 4 {
			obReq.BuyingAssetType = horizonclient.AssetType12
		}
	}

	obReq.Limit = 1

	ob, err := client.OrderBook(obReq)
	if err == nil {
		if len(ob.Bids) > 0 {
			data.BestBid = ob.Bids[0].Price
		}
		if len(ob.Asks) > 0 {
			data.BestAsk = ob.Asks[0].Price
		}
	}

	// Try to find liquidity pool
	pairKey := fmt.Sprintf("%s-%s", assetA.GetCode(), assetB.GetCode())
	reversePairKey := fmt.Sprintf("%s-%s", assetB.GetCode(), assetA.GetCode())

	poolID := ""
	if id, ok := poolIDsMap[pairKey]; ok {
		poolID = id
	} else if id, ok := poolIDsMap[reversePairKey]; ok {
		poolID = id
	}

	// Note: LP data is fetched from stellar.expert API in the main app
	// We just store the pool ID if found
	if poolID != "" {
		data.LPPoolID = poolID
		// Could fetch from stellar.expert here if needed
		data.LPLockedA = "--"
		data.LPLockedB = "--"
	}

	return data, nil
}
