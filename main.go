package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/EpiK-Protocol/epik-wallet-golib/epik"
	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
)

var nodeURL = "ws://120.55.82.202:1234/rpc/v0"

var token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJBbGxvdyI6WyJyZWFkIl19.Oivn9CdQ_kB4TriYfoo2CzWQBeCbj9FbH2hVu4ogyBI"

func main() {
	seed, err := hdwallet.NewSeed()
	panicErr(err)
	fmt.Printf("seed:%x\n", seed)
	wallet, err := hdwallet.NewFromSeed(seed)
	panicErr(err)
	path := hdwallet.MustParseDerivationPath("m/44'/60'/0'/0/0")
	account, err := wallet.Derive(path, false)
	panicErr(err)
	erc20Address := account.Address.Hex()
	fmt.Println(erc20Address)
	weixin := "18901085780"
	hash := sha256.Sum256([]byte(weixin))
	erc20Sign, err := wallet.SignHash(account, hash[:])
	panicErr(err)
	fmt.Printf("sign:%x\n", erc20Sign)

	epikWallet, err := epik.NewWallet()
	panicErr(err)
	epikAddr, err := epikWallet.GenerateKey("bls", seed, "m/44'/196'/1'/0/0")
	panicErr(err)
	fmt.Printf("epik addr:%s\n", epikAddr)
	epikSign, err := epikWallet.Sign(epikAddr, hash[:])
	panicErr(err)
	fmt.Printf("epikSign:%x\n", epikSign)
	body := &struct {
		Weixin         string `json:"weixin"`
		EpikAddress    string `json:"epik_address"`
		Erc20Address   string `json:"erc20_address"`
		EpikSignature  string `json:"epik_signature"`
		Erc20Signature string `json:"erc20_signature"`
	}{
		Weixin:         weixin,
		EpikAddress:    epikAddr,
		Erc20Address:   erc20Address,
		EpikSignature:  fmt.Sprintf("%x", epikSign),
		Erc20Signature: fmt.Sprintf("%x", erc20Sign),
	}
	b, err := json.Marshal(body)
	resp, err := http.Post("http://127.0.0.1:3002/testnet/signup", "application/json", bytes.NewReader(b))
	panicErr(err)
	respbody := []byte{}
	_, err = resp.Body.Read(respbody)
	fmt.Println(resp.StatusCode, string(respbody))
}

func panicErr(err error) {
	if err != nil {
		panic(err)
	}
}
