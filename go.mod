module github.com/EpiK-Protocol/epik-wallet-golib

go 1.14

require (
	github.com/EpiK-Protocol/go-epik v0.4.2-0.20200901164337-9b91560725db
	github.com/ethereum/go-ethereum v1.9.20
	github.com/filecoin-project/go-address v0.0.2-0.20200504173055-8b6f2fb2b3ef
	github.com/filecoin-project/specs-actors v0.6.2-0.20200724193152-534b25bdca30
	github.com/ipfs/go-cid v0.0.6
	github.com/miguelmota/go-ethereum-hdwallet v0.0.0-20200123000308-a60dcd172b4c
	github.com/shopspring/decimal v1.2.0
	github.com/supranational/blst v0.1.2-alpha.1.0.20200829171259-c3ee69d4da5b // indirect
)

replace github.com/filecoin-project/specs-actors => ../go-epik-actors

replace github.com/ethereum/go-ethereum => ./extern/go-ethereum
