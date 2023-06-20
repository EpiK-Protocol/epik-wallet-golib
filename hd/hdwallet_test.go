package hd

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

var wallet *Wallet

func init() {
	SetDebug(true)
	var err error
	wallet, err = NewFromMnemonic("fine bubble drum remember motor kiss arctic leisure adjust immune involve expect")
	panicErr(err)
	err = wallet.SetRPC("wss://ropsten.infura.io/ws/v3/1bbd25bd3af94ca2b294f93c346f69cd")
	panicErr(err)
	address, err := wallet.Derive("m/44'/60'/0'/0/0", true)
	panicErr(err)
	fmt.Printf("address:%s\n", address)
	bu, err := wallet.TokenBalance(address, "USDT")
	panicErr(err)
	fmt.Printf("USDT:	%s\n", bu)
	bepk, err := wallet.TokenBalance(address, "EPK")
	panicErr(err)
	fmt.Printf("EPK:	%s\n", bepk)
	buni, err := wallet.TokenBalance(address, "UNI")
	panicErr(err)
	fmt.Printf("UNI:	%s\n", buni)
	beth, err := wallet.Balance(address)
	panicErr(err)
	fmt.Printf("ETH:	%s\n", beth)
}

func TestBalance(t *testing.T) {
	bal, err := wallet.TokenBalance("0x9708D53A5080c66B96c0AdfEf0255EB43564908E", "USDT")
	panicErr(err)
	fmt.Println(bal)
}

func panicErr(err error) {
	if err != nil {
		panic(err)
	}
}

func TestUniswapAmountIn(t *testing.T) {
	amts, _ := wallet.UniswapGetAmountsOut("USDT", "EPK", "0.1")
	fmt.Println("aminajust:", amts.AmountIn, ";", "amout:", amts.AmountOut)
}

func TestUniswapUSDTtoEPK(t *testing.T) {
	hash, err := wallet.UniswapExactTokenForTokens("0x0FdFC04e8c49cdFfA5A69278BAC26E70E79DcB35", "USDT", "EPK", "100", "90", fmt.Sprintf("%d", time.Now().Add(time.Hour*2).Unix()))
	panicErr(err)
	t.Log("hash:", hash)
}

func TestUniswapEPKtoUSDT(t *testing.T) {
	hash, err := wallet.UniswapExactTokenForTokens("0x0FdFC04e8c49cdFfA5A69278BAC26E70E79DcB35", "EPK", "USDT", "2", "0.09", fmt.Sprintf("%d", time.Now().Add(time.Hour*2).Unix()))
	panicErr(err)
	t.Log("hash:", hash)
}

func TestUniswapAddLiquidity(t *testing.T) {
	hash, err := wallet.UniswapAddLiquidity("0x0FdFC04e8c49cdFfA5A69278BAC26E70E79DcB35", "USDT", "EPK", "100", "100", "90", "90", fmt.Sprintf("%d", time.Now().Add(time.Hour*2).Unix()))
	panicErr(err)
	t.Log("hash:", hash)
}

func TestUniswapRemoveLiquidity(t *testing.T) {
	hash, err := wallet.UniswapRemoveLiquidity("0x0FdFC04e8c49cdFfA5A69278BAC26E70E79DcB35", "USDT", "EPK", "99.990000999900009998", "90", "90", fmt.Sprintf("%d", time.Now().Add(time.Hour*2).Unix()))
	panicErr(err)
	t.Log("hash:", hash)
}

func TestLiquidityInfo(t *testing.T) {
	info, err := wallet.UniswapInfo("0x0FdFC04e8c49cdFfA5A69278BAC26E70E79DcB35")
	panicErr(err)
	data, _ := json.Marshal(info)
	t.Log(string(data))
}

func TestVerifyTX(t *testing.T) {
	client, err := ethclient.DialContext(context.Background(), "wss://ropsten.infura.io/ws/v3/1bbd25bd3af94ca2b294f93c346f69cd")
	panicErr(err)
	defer client.Close()
	hash := common.HexToHash("0x2394649837d3b847c8067bfca86b477b053efd68153532098d50c3bddfff4c38")
	_, isPending, err := client.TransactionByHash(context.Background(), hash)
	panicErr(err)
	fmt.Println(isPending)
	receipt, err := client.TransactionReceipt(context.Background(), hash)
	panicErr(err)
	fmt.Println(receipt)
}

func TestAccelerateTx(t *testing.T) {
	addr := "0x0FdFC04e8c49cdFfA5A69278BAC26E70E79DcB35"
	// txHash, err := wallet.TransferToken(addr, "0x093f9569dF5Fa34c3AeC2E855D264f6a140c642e", "USDT", "1.1")
	txHash, err := wallet.Transfer(addr, "0x093f9569dF5Fa34c3AeC2E855D264f6a140c642e", "0.0001")
	panicErr(err)
	fmt.Printf("First Tx Hash:	%s\n", txHash)
	txHash, err = wallet.AccelerateTx(txHash, 1.2)
	panicErr(err)
	fmt.Printf("Secend Tx Hash:	%s\n", txHash)
}

func TestCancelTx(t *testing.T) {
	addr := "0x0FdFC04e8c49cdFfA5A69278BAC26E70E79DcB35"
	txHash, err := wallet.Transfer(addr, "0x093f9569dF5Fa34c3AeC2E855D264f6a140c642e", "0.00011")
	panicErr(err)
	fmt.Printf("First Tx Hash:	%s\n", txHash)
	txHash, err = wallet.CancelTx(txHash)
	panicErr(err)
	fmt.Printf("Secend Tx Hash:	%s\n", txHash)
}
