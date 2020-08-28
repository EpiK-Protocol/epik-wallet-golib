module github.com/EpiK-Protocol/epik-wallet-golib

go 1.14

require (
	github.com/EpiK-Protocol/go-epik v0.4.2-0.20200826134507-7347cf70b04d
	github.com/davidlazar/go-crypto v0.0.0-20200604182044-b73af7476f6c // indirect
	github.com/filecoin-project/go-address v0.0.2-0.20200504173055-8b6f2fb2b3ef
	github.com/filecoin-project/go-crypto v0.0.0-20191218222705-effae4ea9f03
	github.com/filecoin-project/go-data-transfer v0.5.2 // indirect
	github.com/filecoin-project/go-fil-markets v0.5.0 // indirect
	github.com/filecoin-project/go-multistore v0.0.3 // indirect
	github.com/filecoin-project/go-statemachine v0.0.0-20200813232949-df9b130df370 // indirect
	github.com/filecoin-project/sector-storage v0.0.0-20200630180318-4c1968f62a8f
	github.com/filecoin-project/specs-actors v0.9.3
	github.com/filecoin-project/specs-storage v0.1.1-0.20200730063404-f7db367e9401 // indirect
	github.com/google/gopacket v1.1.18 // indirect
	github.com/hannahhoward/cbor-gen-for v0.0.0-20200723175505-5892b522820a // indirect
	github.com/ipfs/go-bitswap v0.2.20 // indirect
	github.com/ipfs/go-block-format v0.0.2
	github.com/ipfs/go-cid v0.0.7
	github.com/ipfs/go-graphsync v0.1.1 // indirect
	github.com/ipfs/go-ipfs-blockstore v1.0.1 // indirect
	github.com/ipfs/go-log/v2 v2.1.2-0.20200626104915-0016c0b4b3e4
	github.com/ipfs/go-merkledag v0.3.2 // indirect
	github.com/libp2p/go-libp2p v0.11.0 // indirect
	github.com/libp2p/go-libp2p-record v0.1.3 // indirect
	github.com/libp2p/go-libp2p-testing v0.1.2-0.20200422005655-8775583591d8 // indirect
	github.com/miguelmota/go-ethereum-hdwallet v0.0.0-20200123000308-a60dcd172b4c // indirect
	github.com/minio/blake2b-simd v0.0.0-20160723061019-3f5f724cb5b1
	github.com/multiformats/go-multihash v0.0.14
	github.com/onsi/ginkgo v1.14.0 // indirect
	github.com/supranational/blst v0.1.2-alpha.1.0.20200827114059-44244b0755b7
	github.com/whyrusleeping/cbor-gen v0.0.0-20200710004633-5379fc63235d
	github.com/xorcare/golden v0.6.1-0.20191112154924-b87f686d7542 // indirect
	go.opencensus.io v0.22.4
	golang.org/x/net v0.0.0-20200707034311-ab3426394381 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1
)

replace github.com/EpiK-Protocol/go-epik => ./extern/go-epik

replace github.com/filecoin-project/specs-actors => ./extern/go-epik-actors

replace github.com/golangci/golangci-lint => github.com/golangci/golangci-lint v1.18.0

replace github.com/filecoin-project/filecoin-ffi => ./extern/go-epik/extern/filecoin-ffi
