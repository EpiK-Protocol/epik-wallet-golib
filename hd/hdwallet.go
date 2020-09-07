package hd

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"net/url"
	"time"

	"github.com/EpiK-Protocol/epik-wallet-golib/abi/epk"
	"github.com/EpiK-Protocol/epik-wallet-golib/abi/usdt"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
	"github.com/shopspring/decimal"
)

//Wallet ...
type Wallet struct {
	hdWallet  *hdwallet.Wallet
	ethClient *ethclient.Client
	rpcURL    string
}

type currencyType string

const (
	USDT currencyType = "USDT"
	EPK  currencyType = "EPK"
)

var contractAddress = map[currencyType]string{
	USDT: "0xdac17f958d2ee523a2206206994597c13d831ec7",
	EPK:  "0xDaF88906aC1DE12bA2b1D2f7bfC94E9638Ac40c4",
}

//NewFromMnemonic ...
func NewFromMnemonic(mnemonic string) (wallet *Wallet, err error) {
	wallet = &Wallet{}
	wallet.hdWallet, err = hdwallet.NewFromMnemonic(mnemonic)
	if err != nil {
		return nil, err
	}
	return
}

//NewFromSeed ...
func NewFromSeed(seed []byte) (wallet *Wallet, err error) {
	wallet = &Wallet{}
	wallet.hdWallet, err = hdwallet.NewFromSeed(seed)
	if err != nil {
		return nil, err
	}
	return
}

//NewMnemonic ...
func NewMnemonic(bits int) (mnemonic string, err error) {
	return hdwallet.NewMnemonic(bits)
}

//SeedFromMnemonic ...
func SeedFromMnemonic(mnemonic string) (seed []byte, err error) {
	return hdwallet.NewSeedFromMnemonic(mnemonic)
}

//NewSeed ...
func NewSeed() (seed []byte, err error) {
	return hdwallet.NewSeed()
}

//SetRPC ...
func (wallet *Wallet) SetRPC(url string) (err error) {
	wallet.rpcURL = url
	return
}

//Accounts ...
func (wallet *Wallet) Accounts() (addrs string) {
	accs := wallet.hdWallet.Accounts()
	addresses := []string{}
	for _, acc := range accs {
		addresses = append(addresses, acc.Address.Hex())
	}
	data, _ := json.Marshal(&addresses)
	return string(data)
}

//Contains ...
func (wallet *Wallet) Contains(address string) bool {
	addr := common.HexToAddress(address)
	account := accounts.Account{Address: addr}
	return wallet.hdWallet.Contains(account)
}

//Derive ...
func (wallet *Wallet) Derive(path string, pin bool) (address string, err error) {
	p, err := hdwallet.ParseDerivationPath(path)
	if err != nil {
		return
	}
	acc, err := wallet.hdWallet.Derive(p, pin)
	if err != nil {
		return
	}
	return acc.Address.Hex(), nil
}

//SignHash ...
func (wallet *Wallet) SignHash(address string, hash []byte) (signature []byte, err error) {
	addr := common.HexToAddress(address)
	account := accounts.Account{Address: addr}
	return wallet.hdWallet.SignHash(account, hash)
}

//SignText ...
func (wallet *Wallet) SignText(address string, text string) (signature []byte, err error) {
	addr := common.HexToAddress(address)
	account := accounts.Account{Address: addr}
	return wallet.hdWallet.SignText(account, []byte(text))
}

//Balance ...
func (wallet *Wallet) Balance(address string) (balance string, err error) {
	if wallet.rpcURL == "" {
		return "", fmt.Errorf("No RPC URL")
	}
	client, err := ethclient.DialContext(context.Background(), wallet.rpcURL)
	if err != nil {
		return
	}
	defer client.Close()
	addr := common.HexToAddress(address)
	bal, err := client.BalanceAt(context.Background(), addr, nil)
	if err != nil {
		return
	}
	balance = BigIntDiv(bal.String(), 18)
	return
}

//TokenBalance ...
func (wallet *Wallet) TokenBalance(address string, currency string) (balance string, err error) {
	if wallet.rpcURL == "" {
		return "", fmt.Errorf("No RPC URL")
	}
	client, err := ethclient.DialContext(context.Background(), wallet.rpcURL)
	if err != nil {
		return
	}
	defer client.Close()
	addr := common.HexToAddress(address)
	switch currencyType(currency) {
	case USDT:
		contract := common.HexToAddress(contractAddress[USDT])
		usdtToken, err := usdt.NewUsdt(contract, client)
		if err != nil {
			return "", err
		}
		opts := &bind.CallOpts{}
		bal, err := usdtToken.BalanceOf(opts, addr)
		if err != nil {
			return "", err
		}
		dec, err := usdtToken.Decimals(opts)
		if err != nil {
			return "", err
		}
		balance = BigIntDiv(bal.String(), int(dec.Int64()))
	case EPK:
		contract := common.HexToAddress(contractAddress[EPK])
		epkToken, err := epk.NewEpk(contract, client)
		if err != nil {
			return "", err
		}
		opts := &bind.CallOpts{}
		bal, err := epkToken.BalanceOf(opts, addr)
		if err != nil {
			return "", err
		}
		dec, err := epkToken.Decimals(opts)
		if err != nil {
			return "", err
		}
		balance = BigIntDiv(bal.String(), int(dec))
	default:
		return "", fmt.Errorf("Currency  Unsuppoted")
	}
	return
}

//Transfer ...
func (wallet *Wallet) Transfer(from string, to string, amount string) (txHash string, err error) {
	if wallet.rpcURL == "" {
		return "", fmt.Errorf("No RPC URL")
	}
	client, err := ethclient.DialContext(context.Background(), wallet.rpcURL)
	if err != nil {
		return
	}
	defer client.Close()
	fromAddr := common.HexToAddress(from)
	toAddr := common.HexToAddress(to)
	amountWei, err := decimal.NewFromString(BigIntMul(amount, 18))
	if err != nil {
		return "", err
	}

	nonce, err := client.NonceAt(context.Background(), fromAddr, nil)
	if err != nil {
		return "", err
	}
	gasLimit := uint64(100000)
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return "", err
	}

	tx := types.NewTransaction(nonce, toAddr, amountWei.BigInt(), gasLimit, gasPrice, nil)
	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		return "", err
	}
	var account accounts.Account
	find := false
	for _, acc := range wallet.hdWallet.Accounts() {
		if acc.Address == fromAddr {
			find = true
			account = acc
			break
		}
	}
	if !find {
		return "", fmt.Errorf("Account Not Found")
	}
	privateKey, err := wallet.hdWallet.PrivateKey(account)
	if err != nil {
		return "", err
	}
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return "", err
	}
	return signedTx.Hash().String(), nil
}

//TransferToken ...
func (wallet *Wallet) TransferToken(from string, to string, currency string, amount string) (txHash string, err error) {
	if wallet.rpcURL == "" {
		return "", fmt.Errorf("No RPC URL")
	}
	client, err := ethclient.DialContext(context.Background(), wallet.rpcURL)
	if err != nil {
		return
	}
	defer client.Close()
	fromAddr := common.HexToAddress(from)
	toAddr := common.HexToAddress(to)
	switch currencyType(currency) {
	case USDT:
		contract := common.HexToAddress(contractAddress[USDT])
		usdtToken, err := usdt.NewUsdt(contract, client)
		if err != nil {
			return "", err
		}
		opts := &bind.CallOpts{}
		bal, err := usdtToken.BalanceOf(opts, fromAddr)
		if err != nil {
			return "", err
		}
		dec, err := usdtToken.Decimals(opts)
		amountWei, err := decimal.NewFromString(BigIntMul(amount, int(dec.Int64())))
		if err != nil {
			return "", err
		}
		if amountWei.Cmp(decimal.NewFromBigInt(bal, 10)) > 0 {
			return "", fmt.Errorf("Out of Balance")
		}
		gasPrice, err := client.SuggestGasPrice(context.Background())
		if err != nil {
			return "", err
		}
		nonce, err := client.PendingNonceAt(context.Background(), fromAddr)
		if err != nil {
			return "", err
		}
		var account accounts.Account
		find := false
		for _, acc := range wallet.hdWallet.Accounts() {
			if acc.Address == fromAddr {
				find = true
				account = acc
				break
			}
		}
		if !find {
			return "", fmt.Errorf("Account Not Found")
		}
		privateKey, err := wallet.hdWallet.PrivateKey(account)
		if err != nil {
			return "", err
		}
		auth := bind.NewKeyedTransactor(privateKey)
		auth.Nonce = big.NewInt(int64(nonce))
		auth.GasLimit = uint64(60000)
		auth.GasPrice = gasPrice
		tx, err := usdtToken.Transfer(auth, toAddr, amountWei.BigInt())
		if err != nil {
			return "", err
		}
		txHash = tx.Hash().String()
	case EPK:
		contract := common.HexToAddress(contractAddress[EPK])
		epkToken, err := epk.NewEpk(contract, client)
		if err != nil {
			return "", err
		}
		opts := &bind.CallOpts{}
		bal, err := epkToken.BalanceOf(opts, fromAddr)
		if err != nil {
			return "", err
		}
		dec, err := epkToken.Decimals(opts)
		amountWei, err := decimal.NewFromString(BigIntMul(amount, int(dec)))
		if err != nil {
			return "", err
		}
		if amountWei.Cmp(decimal.NewFromBigInt(bal, 10)) > 0 {
			return "", fmt.Errorf("Out of Balance")
		}
		gasPrice, err := client.SuggestGasPrice(context.Background())
		if err != nil {
			return "", err
		}
		nonce, err := client.PendingNonceAt(context.Background(), fromAddr)
		if err != nil {
			return "", err
		}
		var account accounts.Account
		find := false
		for _, acc := range wallet.hdWallet.Accounts() {
			if acc.Address == fromAddr {
				find = true
				account = acc
				break
			}
		}
		if !find {
			return "", fmt.Errorf("Account Not Found")
		}
		privateKey, err := wallet.hdWallet.PrivateKey(account)
		if err != nil {
			return "", err
		}
		auth := bind.NewKeyedTransactor(privateKey)
		auth.Nonce = big.NewInt(int64(nonce))
		auth.GasLimit = uint64(60000)
		auth.GasPrice = gasPrice

		tx, err := epkToken.Transfer(auth, toAddr, amountWei.BigInt())
		if err != nil {
			return "", err
		}
		txHash = tx.Hash().String()
	default:
		return "", fmt.Errorf("Currency  Unsuppoted")
	}
	return
}

var httpClient = &http.Client{Timeout: time.Duration(20 * time.Second)}

//Transactions ...
func (wallet *Wallet) Transactions(address string, currency string, page, offset int64, asc bool) (txs string, err error) {
	u, _ := url.Parse("https://tx.epik-protocol.io/api")
	query := u.Query()
	query.Set("module", "account")
	query.Set("page", fmt.Sprintf("%d", page))
	query.Set("offset", fmt.Sprintf("%d", offset))
	if asc {
		query.Set("sort", "asc")
	} else {
	}
	switch currencyType(currency) {
	case USDT, EPK:
		query.Set("action", "tokentx")
	default:
		query.Set("action", "txlist")
	}
	query.Set("address", address)
	query.Set("contractaddress", contractAddress[currencyType(currency)])
	u.RawQuery = query.Encode()
	resp, err := http.Get(u.String())
	if err != nil {
		return
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf(resp.Status)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	txs = string(body)
	return
}

func BigIntDiv(balance string, decimals int) string {
	bal, _ := decimal.NewFromString(balance)
	for i := 0; i < decimals; i++ {
		bal = bal.Div(decimal.NewFromInt(10))
	}
	return bal.String()
}
func BigIntMul(balance string, decimals int) string {
	bal, _ := decimal.NewFromString(balance)
	for i := 0; i < decimals; i++ {
		bal = bal.Mul(decimal.NewFromInt(10))
	}
	return bal.String()
}
