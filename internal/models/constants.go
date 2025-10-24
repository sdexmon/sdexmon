package models

import (
	"time"

	"github.com/stellar/go/txnbuild"
)

const (
	DefaultDepth      = 7 // rows per side (limit to 7 each side)
	OrderbookInterval = 1200 * time.Millisecond
	TradesInterval    = 1200 * time.Millisecond
	LPInterval        = 30 * time.Second
	MaxTradesKept     = 120
)

const ASCIIAquila = `███████  ██████  █████  ██████       █████   ██████  ██    ██ ██ ██       █████  
██      ██      ██   ██ ██   ██     ██   ██ ██    ██ ██    ██ ██ ██      ██   ██ 
███████ ██      ███████ ██████      ███████ ██    ██ ██    ██ ██ ██      ███████ 
     ██ ██      ██   ██ ██   ██     ██   ██ ██ ▄▄ ██ ██    ██ ██ ██      ██   ██ 
███████  ██████ ██   ██ ██   ██     ██   ██  ██████   ██████  ██ ███████ ██   ██ 
                                                ▀▀                               
                                                                                 `

// CuratedAssets are the predefined assets for the TUI
var CuratedAssets = map[string]txnbuild.Asset{
	"USDZ": txnbuild.CreditAsset{Code: "USDZ", Issuer: "GAKTLPC4ZV37SSCITQ5IS5AQ4WPF4CF4VZJQPPAROSGXMYOATF5U6XPR"},
	"ZARZ": txnbuild.CreditAsset{Code: "ZARZ", Issuer: "GAROH4EV3WVVTRQKEY43GZK3XSRBEYETRVZ7SVG5LHWOAANSMCTJBB3U"},
	"EURZ": txnbuild.CreditAsset{Code: "EURZ", Issuer: "GAM5BKSKTHYS6IE4OUHCISGI6YVH75XIMOCG4RB5TR74KZDJRSNKEURZ"},
	"XAUZ": txnbuild.CreditAsset{Code: "XAUZ", Issuer: "GD3MMNHD5U5H732GTLYO7DZVUNGPVP462KVNFO4HALNPP6C7ESQAGOLD"},
	"BTCZ": txnbuild.CreditAsset{Code: "BTCZ", Issuer: "GAT63G6FINKAES4473ZZZT3SYJVUIXKYBVFBQYQHEZF6EE3VY5AGBTCZ"},
	"XLM":  txnbuild.NativeAsset{},
	"USDC": txnbuild.CreditAsset{Code: "USDC", Issuer: "GA5ZSEJYB37JRC5AVCIA5MOP4RHTM335X2KGX3IHOJAPP5RE34K4KZVN"},
}

// CuratedPairs are the predefined trading pairs
var CuratedPairs = []PairOption{
	{"USDC", "USDZ"},
	{"USDZ", "ZARZ"},
	{"USDZ", "EURZ"},
	{"USDZ", "BTCZ"},
	{"USDZ", "XAUZ"},
	{"EURZ", "ZARZ"},
	{"EURZ", "XAUZ"},
	{"EURZ", "BTCZ"},
	{"ZARZ", "XAUZ"},
	{"ZARZ", "BTCZ"},
	{"XAUZ", "BTCZ"},
	{"XLM", "USDC"},
	{"XLM", "USDZ"},
	{"XLM", "EURZ"},
	{"XLM", "ZARZ"},
	{"XLM", "XAUZ"},
	{"XLM", "BTCZ"},
}

// LiquidityPoolIDs maps pair keys to pool IDs
var LiquidityPoolIDs = map[string]string{
	"USDC-USDZ": "314e17d86ffc767a6132fba31cc9f53f23ca359d2db788f26f0d364d75e82c57",
	"USDZ-USDC": "314e17d86ffc767a6132fba31cc9f53f23ca359d2db788f26f0d364d75e82c57",
	"USDZ-ZARZ": "d6842cf8f10ec2fc8a4599f23f7b0161bafa228b267714fc3ed6ca6d48d0b13c",
	"ZARZ-USDZ": "d6842cf8f10ec2fc8a4599f23f7b0161bafa228b267714fc3ed6ca6d48d0b13c",
	"USDZ-EURZ": "30869ce3dd1e130649c08ca0986bcb912acd4c557502378d8e32f41e1c443f55",
	"EURZ-USDZ": "30869ce3dd1e130649c08ca0986bcb912acd4c557502378d8e32f41e1c443f55",
	"USDZ-BTCZ": "645923faa8b51f09f63306db95788bf4d8aa033ff50031ac279dcdb483207f10",
	"BTCZ-USDZ": "645923faa8b51f09f63306db95788bf4d8aa033ff50031ac279dcdb483207f10",
	"USDZ-XAUZ": "f0344bb1fbde3157c745ca7c310e9516877ef30ae35cacf3f268b4b163d30788",
	"XAUZ-USDZ": "f0344bb1fbde3157c745ca7c310e9516877ef30ae35cacf3f268b4b163d30788",
	"EURZ-ZARZ": "57b50011b18e2e6a94b4cf745a569779a50d710c757caa37d38148d24d383cc9",
	"ZARZ-EURZ": "57b50011b18e2e6a94b4cf745a569779a50d710c757caa37d38148d24d383cc9",
	"EURZ-XAUZ": "1c473914c3af39f5ed04284f01f8488906ec9ddeae31e3f4dc608e9871ba4a68",
	"XAUZ-EURZ": "1c473914c3af39f5ed04284f01f8488906ec9ddeae31e3f4dc608e9871ba4a68",
	"EURZ-BTCZ": "3c3d8532451361b47986d1c864e029488453fcf923bca383af673a4fe84ef8c1",
	"BTCZ-EURZ": "3c3d8532451361b47986d1c864e029488453fcf923bca383af673a4fe84ef8c1",
	"ZARZ-XAUZ": "962528fd96913f256044daf4aa77162be04c381764fef6f92b6962b4d6c50fb1",
	"XAUZ-ZARZ": "962528fd96913f256044daf4aa77162be04c381764fef6f92b6962b4d6c50fb1",
	"ZARZ-BTCZ": "39b4a2889462d58dffb9e11a97502f2a74788d9c2b6c6b711ba2e7b0cfe2a7d8",
	"BTCZ-ZARZ": "39b4a2889462d58dffb9e11a97502f2a74788d9c2b6c6b711ba2e7b0cfe2a7d8",
	"XAUZ-BTCZ": "a4753a9faa6b256e46fb63ab900c64333d5d799ee48b70452d3fa833db350f33",
	"BTCZ-XAUZ": "a4753a9faa6b256e46fb63ab900c64333d5d799ee48b70452d3fa833db350f33",
	"XLM-USDC":  "a468d41d8e9b8f3c7209651608b74b7db7ac9952dcae0cdf24871d1d9c7b0088",
	"USDC-XLM":  "a468d41d8e9b8f3c7209651608b74b7db7ac9952dcae0cdf24871d1d9c7b0088",
	"XLM-USDZ":  "7001fca2d71456cda8a061e4733f035fce36423ccf942e92db139a116d7e557b",
	"USDZ-XLM":  "7001fca2d71456cda8a061e4733f035fce36423ccf942e92db139a116d7e557b",
	"XLM-EURZ":  "d79c741bc6371240af4a1e86c645742a561581095bd147ae86a0a3386701c545",
	"EURZ-XLM":  "d79c741bc6371240af4a1e86c645742a561581095bd147ae86a0a3386701c545",
	"XLM-ZARZ":  "fb7072d551e853826e4a5497e2da1e6765c8cc29fa938ceeeeef579adc53a9f6",
	"ZARZ-XLM":  "fb7072d551e853826e4a5497e2da1e6765c8cc29fa938ceeeeef579adc53a9f6",
	"XLM-XAUZ":  "fb0e4a67424a2851cfa02618de758f2cbaa71e737454caf25919fa51bab125e5",
	"XAUZ-XLM":  "fb0e4a67424a2851cfa02618de758f2cbaa71e737454caf25919fa51bab125e5",
	"XLM-BTCZ":  "d8905565dac7e4c5702520bdf39d1e8b385a94708628c87333862a41b62da980",
	"BTCZ-XLM":  "d8905565dac7e4c5702520bdf39d1e8b385a94708628c87333862a41b62da980",
}
