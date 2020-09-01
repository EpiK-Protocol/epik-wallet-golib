package hd

import (
	"fmt"
	"testing"
)

func TestBalance(t *testing.T) {
	mnen, err := NewMnemonic(128)
	panicErr(err)
	wallet, err := NewFromMnemonic(mnen)
	panicErr(err)
	err = wallet.SetRPC("https://mainnet.infura.io/v3/1bbd25bd3af94ca2b294f93c346f69cd")
	panicErr(err)
	balance, err := wallet.Balance("0x010C08D59Be466F6e7800Ec7eC97397160971F64")
	panicErr(err)
	fmt.Println(balance)
}

func panicErr(err error) {
	if err != nil {
		panic(err)
	}
}
