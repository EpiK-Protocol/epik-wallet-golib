package epik

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
)

func TestGenKey(t *testing.T) {
	// addr, err := GenerateKey("bls")
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Printf("GENKEY:%s\n", addr)
}

func TestSignupEpik(t *testing.T) {
	mnec := "truth later fuel float various bright demise surprise two plunge minor cram"
	fmt.Println(mnec)
	seed, err := hdwallet.NewSeedFromMnemonic(mnec)
	panicErr(err)
	fmt.Printf("seed:%x\n", seed)
	wallet, err := hdwallet.NewFromSeed(seed)
	panicErr(err)
	path := hdwallet.MustParseDerivationPath("m/44'/60'/0'/0/0")
	account, err := wallet.Derive(path, true)
	panicErr(err)
	fmt.Println(wallet.Accounts())
	erc20Address := account.Address.Hex()
	fmt.Println(erc20Address)
	weixin := "18901085780"
	hash := sha256.Sum256([]byte(weixin))
	erc20Sign, err := wallet.SignHash(account, hash[:])
	erc20SignStr := hex.EncodeToString(erc20Sign)
	panicErr(err)
	fmt.Printf("sign:%s\n", erc20SignStr)

	epikWallet, err := NewWallet()
	panicErr(err)
	epikAddr, err := epikWallet.GenerateKey("bls", seed[0:32])
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
		Erc20Signature: erc20SignStr,
	}
	b, err := json.Marshal(body)
	resp, err := http.Post("http://127.0.0.1:3002/testnet/signup", "application/json", bytes.NewReader(b))
	panicErr(err)
	respbody, err := ioutil.ReadAll(resp.Body)
	fmt.Println(resp.StatusCode, string(respbody))
}

func panicErr(err error) {
	if err != nil {
		panic(err)
	}
}
