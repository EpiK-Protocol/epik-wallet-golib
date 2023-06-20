module github.com/EpiK-Protocol/epik-wallet-golib

go 1.16

require (
	github.com/EpiK-Protocol/go-epik v0.4.2-0.20200901164337-9b91560725db
	github.com/btcsuite/btcd v0.22.0-beta
	github.com/btcsuite/btcutil v1.0.3-0.20201208143702-a53e38424cce
	github.com/ethereum/go-ethereum v1.10.8
	github.com/filecoin-project/go-address v0.0.5
	github.com/filecoin-project/go-fil-commcid v0.1.0 // indirect
	github.com/filecoin-project/go-jsonrpc v0.1.4-0.20210217175800-45ea43ac2bec
	github.com/filecoin-project/go-state-types v0.1.0
	github.com/filecoin-project/specs-actors/v2 v2.3.4
	github.com/golang/mock v1.5.0 // indirect
	github.com/hashicorp/go-uuid v1.0.1
	github.com/ipfs/go-cid v0.0.7
	github.com/ipfs/go-log/v2 v2.1.2-0.20200626104915-0016c0b4b3e4
	github.com/klauspost/cpuid/v2 v2.0.6 // indirect
	github.com/miguelmota/go-ethereum-hdwallet v0.1.1
	github.com/multiformats/go-multihash v0.0.15 // indirect
	github.com/polydawn/refmt v0.0.0-20201211092308-30ac6d18308e // indirect
	github.com/rjeczalik/notify v0.9.2 // indirect
	github.com/shopspring/decimal v1.2.0
	github.com/supranational/blst v0.2.0
	github.com/tyler-smith/go-bip39 v1.1.0
	github.com/whyrusleeping/cbor-gen v0.0.0-20210303213153-67a261a1d291 // indirect
	github.com/xlab/c-for-go v0.0.0-20201223145653-3ba5db515dcb // indirect
	golang.org/x/crypto v0.0.0-20210817164053-32db794688a5 // indirect
	golang.org/x/lint v0.0.0-20201208152925-83fdc39ff7b5 // indirect
	golang.org/x/mobile v0.0.0-20220722155234-aaac322e2105 // indirect
	golang.org/x/sys v0.0.0-20210902050250-f475640dd07b // indirect
	golang.org/x/tools v0.1.8-0.20211022200916-316ba0b74098 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1
	google.golang.org/grpc v1.36.0 // indirect
	modernc.org/cc v1.0.1 // indirect
	modernc.org/strutil v1.1.1 // indirect

)

replace github.com/ethereum/go-ethereum => ../go-ethereum

replace github.com/filecoin-project/specs-actors/v2 => github.com/EpiK-Protocol/go-epik-actors/v2 v2.4.0-alpha.0.20211007091141-d6b2892aaedc

replace github.com/EpiK-Protocol/go-epik => ../go-epik
