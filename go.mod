module github.com/EpiK-Protocol/epik-wallet-golib

go 1.15

require (
	github.com/EpiK-Protocol/go-epik v0.4.2-0.20200901164337-9b91560725db
	github.com/btcsuite/btcd v0.20.1-beta
	github.com/btcsuite/btcutil v0.0.0-20190425235716-9e5f4b9a998d
	github.com/ethereum/go-ethereum v1.9.20
	github.com/filecoin-project/go-address v0.0.5
	github.com/filecoin-project/go-fil-commcid v0.1.0 // indirect
	github.com/filecoin-project/go-jsonrpc v0.1.4-0.20210217175800-45ea43ac2bec
	github.com/filecoin-project/go-state-types v0.1.0
	github.com/filecoin-project/specs-actors/v2 v2.3.4
	github.com/golang/mock v1.5.0 // indirect
	github.com/golang/protobuf v1.5.1 // indirect
	github.com/hashicorp/go-uuid v1.0.1
	github.com/ipfs/go-cid v0.0.7
	github.com/ipfs/go-log/v2 v2.1.2-0.20200626104915-0016c0b4b3e4
	github.com/klauspost/cpuid/v2 v2.0.6 // indirect
	github.com/miguelmota/go-ethereum-hdwallet v0.0.0-20200123000308-a60dcd172b4c
	github.com/multiformats/go-multihash v0.0.15 // indirect
	github.com/polydawn/refmt v0.0.0-20201211092308-30ac6d18308e // indirect
	github.com/shopspring/decimal v1.2.0
	github.com/supranational/blst v0.2.0
	github.com/tyler-smith/go-bip39 v1.0.1-0.20181017060643-dbb3b84ba2ef
	github.com/whyrusleeping/cbor-gen v0.0.0-20210303213153-67a261a1d291 // indirect
	github.com/xlab/c-for-go v0.0.0-20201223145653-3ba5db515dcb // indirect
	go.opencensus.io v0.23.0 // indirect
	golang.org/x/lint v0.0.0-20201208152925-83fdc39ff7b5 // indirect
	golang.org/x/mobile v0.0.0-20210527171505-7e972142eb43 // indirect
	golang.org/x/text v0.3.5 // indirect
	golang.org/x/tools v0.1.2 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1
	google.golang.org/grpc v1.36.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	modernc.org/cc v1.0.1 // indirect
	modernc.org/strutil v1.1.1 // indirect
)

replace github.com/ethereum/go-ethereum => ../go-ethereum

replace github.com/filecoin-project/specs-actors/v2 => github.com/EpiK-Protocol/go-epik-actors/v2 v2.4.0-alpha.0.20210512183918-a7df61bb4cd5

replace github.com/EpiK-Protocol/go-epik => ../go-epik
