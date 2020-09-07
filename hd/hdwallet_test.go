package hd

import (
	"testing"
)

func TestBalance(t *testing.T) {

	wallet, err := NewFromMnemonic("fine bubble drum remember motor kiss arctic leisure adjust immune involve expect")
	panicErr(err)
	err = wallet.SetRPC("https://mainnet.infura.io/v3/1bbd25bd3af94ca2b294f93c346f69cd")
	panicErr(err)
	address, err := wallet.Derive("m/44'/60'/0'/0/0", true)
	panicErr(err)
	t.Logf("address:%s\n", address)
	bu, err := wallet.TokenBalance(address, "USDT")
	panicErr(err)
	t.Logf("USDT:%s\n", bu)
	bepk, err := wallet.TokenBalance(address, "EPK")
	panicErr(err)
	t.Logf("EPK:%s\n", bepk)
	beth, err := wallet.Balance(address)
	panicErr(err)
	t.Logf("ETH:%s\n", beth)
	txs, err := wallet.Transactions(address, "ETH", 0, 30, false)
	panicErr(err)
	t.Logf("TXS:%s\n", txs)
	txhash, err := wallet.TransferToken(address, "0xf1c7b91ec6bd7e72a39feafa33deb748b87cfb12", "EPK", "99")
	panicErr(err)
	t.Logf("HASH:%s\n", txhash)
}

func panicErr(err error) {
	if err != nil {
		panic(err)
	}
}
