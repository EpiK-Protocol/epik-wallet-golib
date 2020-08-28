package main

import (
	"fmt"

	"github.com/EpiK-Protocol/epik-wallet-golib/eth"
)

var nodeURL = "ws://120.55.82.202:1234/rpc/v0"

var token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJBbGxvdyI6WyJyZWFkIl19.Oivn9CdQ_kB4TriYfoo2CzWQBeCbj9FbH2hVu4ogyBI"

func main() {
	entropy, err := eth.NewMnemonic(128)
	fmt.Println(entropy, err)
}
