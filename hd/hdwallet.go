package hd

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"net/url"
	"regexp"

	"github.com/EpiK-Protocol/epik-wallet-golib/abi/epk"
	"github.com/EpiK-Protocol/epik-wallet-golib/abi/uniswap"
	"github.com/EpiK-Protocol/epik-wallet-golib/abi/univ2"
	"github.com/EpiK-Protocol/epik-wallet-golib/abi/usdt"
	"github.com/tyler-smith/go-bip39"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
	"github.com/shopspring/decimal"
)

//Wallet ...
type Wallet struct {
	hdWallet *hdwallet.Wallet
	rpcURL   string
}

type currencyType string

const (
	USDT currencyType = "USDT"
	EPK  currencyType = "EPK"
	UNI  currencyType = "UNI"
)

// //mainnet
var contractAddress = map[currencyType]string{
	USDT: "0xdAC17F958D2ee523a2206206994597C13D831ec7",
	EPK:  "0xDaF88906aC1DE12bA2b1D2f7bfC94E9638Ac40c4",
	UNI:  "0x66e32d1B776a43935ed20E8dc39A27A140d31C8f",
}

var uniswapContract = "0x7a250d5630B4cF539739dF2C5dAcb4c659F2488D" //mainnet
var txHost = "http://tx.epik-protocol.io"

//ropsten
// var contractAddress = map[currencyType]string{
// 	USDT: "0xD28d251684085Af5eCd283847E4666f80094e26B",
// 	EPK:  "0x6936bae5b97c6eba746932e9cfa33931963cd333",
// 	UNI:  "0xfc2d0b1db58fe2b94b5cf0e2604471a2cb2432ca",
// }

// const uniswapContract = "0x5C69bEe701ef814a2B6a3EDD4B1652CB9cc5aA6f" //ropsten
// const txHost = "http://tx-ropsten.epik-protocol.io"

func init() {
	decimal.DivisionPrecision = 18
}

func SetDebug(debug bool) {
	if debug {
		contractAddress = map[currencyType]string{
			USDT: "0xD28d251684085Af5eCd283847E4666f80094e26B",
			EPK:  "0x6936bae5b97c6eba746932e9cfa33931963cd333",
			UNI:  "0xfc2d0b1db58fe2b94b5cf0e2604471a2cb2432ca",
		}
		uniswapContract = "0x5C69bEe701ef814a2B6a3EDD4B1652CB9cc5aA6f" //ropsten
		txHost = "http://tx-ropsten.epik-protocol.io"
	} else {
		contractAddress = map[currencyType]string{
			USDT: "0xdAC17F958D2ee523a2206206994597C13D831ec7",
			EPK:  "0xDaF88906aC1DE12bA2b1D2f7bfC94E9638Ac40c4",
			UNI:  "0x66e32d1B776a43935ed20E8dc39A27A140d31C8f",
		}
		uniswapContract = "0x7a250d5630B4cF539739dF2C5dAcb4c659F2488D" //mainnet
		txHost = "http://tx.epik-protocol.io"

	}
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
	if !bip39.IsMnemonicValid(mnemonic) {
		return nil, errors.New("mnemonic is invalid")
	}
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

func (wallet *Wallet) Export(address string) (privateKey string, err error) {
	addr := common.HexToAddress(address)
	pk, err := wallet.getPrivateKey(addr)
	if err != nil {
		return
	}
	key := crypto.FromECDSA(pk)
	privateKey = hex.EncodeToString(key)
	return
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
	balance = decimal.NewFromBigInt(bal, -18).String()
	return
}

//SuggestGas ...
func (wallet *Wallet) SuggestGas() (gas string, err error) {
	client, err := ethclient.DialContext(context.Background(), wallet.rpcURL)
	if err != nil {
		return
	}
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return "", err
	}
	return decimal.NewFromBigInt(gasPrice, -18).Mul(decimal.NewFromInt(50000)).String(), err
}

//SuggestGasPrice ...
func (wallet *Wallet) SuggestGasPrice() (gasPrice string, err error) {
	client, err := ethclient.DialContext(context.Background(), wallet.rpcURL)
	if err != nil {
		return
	}
	price, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return "", err
	}
	return decimal.NewFromBigInt(price, -18).String(), err
}

//TokenBalance ...
func (wallet *Wallet) TokenBalance(address string, currency string) (balance string, err error) {
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
		balance = decimal.NewFromBigInt(bal, -int32(dec.Int64())).String()
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
		balance = decimal.NewFromBigInt(bal, -int32(dec)).String()
	case UNI:
		contract := common.HexToAddress(contractAddress[UNI])
		uniToken, err := univ2.NewUniv2(contract, client)
		if err != nil {
			return "", err
		}
		opts := &bind.CallOpts{}
		bal, err := uniToken.BalanceOf(opts, addr)
		if err != nil {
			return "", err
		}
		dec, err := uniToken.Decimals(opts)
		if err != nil {
			return "", err
		}
		balance = decimal.NewFromBigInt(bal, -int32(dec)).String()
	default:
		return "", fmt.Errorf("currency  unsuppoted")
	}
	return
}

func checkAddress(address string) bool {
	re := regexp.MustCompile("^0x[0-9a-fA-F]{40}$")
	return re.MatchString(address)
}

//Transfer ...
func (wallet *Wallet) Transfer(from string, to string, amount string) (txHash string, err error) {
	if !checkAddress(from) || !checkAddress(to) {
		return "", fmt.Errorf("address error")
	}
	client, err := ethclient.DialContext(context.Background(), wallet.rpcURL)
	if err != nil {
		return
	}
	defer client.Close()
	fromAddr := common.HexToAddress(from)
	toAddr := common.HexToAddress(to)
	amountWei, err := decimal.NewFromString(amount)
	if err != nil {
		return "", err
	}
	amountWei = amountWei.Mul(decimal.NewFromBigInt(big.NewInt(1), 18))
	nonce, err := client.PendingNonceAt(context.Background(), fromAddr)
	if err != nil {
		return "", err
	}
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return "", err
	}
	gasPrice = new(big.Int).Add(gasPrice, new(big.Int).Div(gasPrice, big.NewInt(10)))
	tx := types.NewTransaction(nonce, toAddr, amountWei.BigInt(), 21000, gasPrice, nil)
	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		return "", err
	}
	privateKey, err := wallet.getPrivateKey(fromAddr)
	if err != nil {
		return "", err
	}
	signer := types.LatestSignerForChainID(chainID)
	signature, err := crypto.Sign(signer.Hash(tx).Bytes(), privateKey)
	if err != nil {
		return "", err
	}
	signedTx, err := tx.WithSignature(signer, signature)
	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return "", err
	}
	return signedTx.Hash().String(), nil
}

//TransferToken ...
func (wallet *Wallet) TransferToken(from string, to string, currency string, amount string) (txHash string, err error) {
	if !checkAddress(from) || !checkAddress(to) {
		return "", fmt.Errorf("address error")
	}
	client, err := ethclient.DialContext(context.Background(), wallet.rpcURL)
	if err != nil {
		return
	}
	defer client.Close()
	chainID, err := client.ChainID(context.Background())
	if err != nil {
		return
	}
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
		if err != nil {
			return "", err
		}
		amountWei, err := decimal.NewFromString(amount)
		if err != nil {
			return "", err
		}
		amountWei = amountWei.Mul(decimal.NewFromBigInt(big.NewInt(1), int32(dec.Int64())))
		if amountWei.Cmp(decimal.NewFromBigInt(bal, 10)) > 0 {
			return "", fmt.Errorf("out of balance")
		}
		privateKey, err := wallet.getPrivateKey(fromAddr)
		if err != nil {
			return "", err
		}
		auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
		if err != nil {
			return "", err
		}
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
		if err != nil {
			return "", err
		}
		amountWei, err := decimal.NewFromString(amount)
		if err != nil {
			return "", err
		}
		amountWei = amountWei.Mul(decimal.NewFromBigInt(big.NewInt(1), int32(dec)))
		if amountWei.Cmp(decimal.NewFromBigInt(bal, 10)) > 0 {
			return "", fmt.Errorf("out of balance")
		}
		privateKey, err := wallet.getPrivateKey(fromAddr)
		if err != nil {
			return "", err
		}
		auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
		if err != nil {
			return "", err
		}
		tx, err := epkToken.Transfer(auth, toAddr, amountWei.BigInt())
		if err != nil {
			return "", err
		}
		txHash = tx.Hash().String()
	case UNI:
		contract := common.HexToAddress(contractAddress[EPK])
		uniToken, err := univ2.NewUniv2(contract, client)
		if err != nil {
			return "", err
		}
		opts := &bind.CallOpts{}
		bal, err := uniToken.BalanceOf(opts, fromAddr)
		if err != nil {
			return "", err
		}
		dec, err := uniToken.Decimals(opts)
		if err != nil {
			return "", err
		}
		amountWei, err := decimal.NewFromString(amount)
		if err != nil {
			return "", err
		}
		amountWei = amountWei.Mul(decimal.NewFromBigInt(big.NewInt(1), int32(dec)))
		if amountWei.Cmp(decimal.NewFromBigInt(bal, 10)) > 0 {
			return "", fmt.Errorf("out of balance")
		}
		privateKey, err := wallet.getPrivateKey(fromAddr)
		if err != nil {
			return "", err
		}
		auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
		if err != nil {
			return "", err
		}
		tx, err := uniToken.Transfer(auth, toAddr, amountWei.BigInt())
		if err != nil {
			return "", err
		}
		txHash = tx.Hash().String()
	default:
		return "", fmt.Errorf("currency  unsuppoted")
	}
	return
}

//Receipt  ...
func (wallet *Wallet) Receipt(txHash string) (status string, err error) {
	client, err := ethclient.DialContext(context.Background(), wallet.rpcURL)
	if err != nil {
		return
	}
	defer client.Close()
	_, isPending, err := client.TransactionByHash(context.Background(), common.HexToHash(txHash))
	if err != nil {
		return
	}
	if isPending {
		return "pending", nil
	}
	receipt, err := client.TransactionReceipt(context.Background(), common.HexToHash(txHash))
	if err != nil {
		return "", err
	}
	switch receipt.Status {
	case types.ReceiptStatusFailed:
		return "failed", nil
	case types.ReceiptStatusSuccessful:
		return "success", nil
	default:
		return "failed", nil
	}
}

//AccelerateTx 加速交易
func (wallet *Wallet) AccelerateTx(srcTxHash string, gasRate float64) (txHash string, err error) {
	client, err := ethclient.DialContext(context.Background(), wallet.rpcURL)
	if err != nil {
		return
	}
	tx, isPending, err := client.TransactionByHash(context.Background(), common.HexToHash(srcTxHash))
	if err != nil {
		return
	}
	if !isPending {
		return "", fmt.Errorf("tx success")
	}
	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		return "", err
	}
	msg, err := tx.AsMessage(types.LatestSignerForChainID(tx.ChainId()), nil)
	if err != nil {
		return
	}
	privateKey, err := wallet.getPrivateKey(msg.From())
	if err != nil {
		return "", err
	}
	gasLimit := uint64(float64(tx.Gas()) * gasRate)
	gasPrice := decimal.NewFromBigInt(tx.GasPrice(), 0).Mul(decimal.NewFromFloat(gasRate))
	tx = types.NewTransaction(tx.Nonce(), *tx.To(), tx.Value(), uint64(float64(gasLimit)*gasRate), gasPrice.BigInt(), tx.Data())

	signer := types.LatestSignerForChainID(chainID)
	signature, err := crypto.Sign(signer.Hash(tx).Bytes(), privateKey)
	if err != nil {
		return "", err
	}
	signedTx, err := tx.WithSignature(signer, signature)
	if err != nil {
		return "", err
	}
	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return "", err
	}
	return signedTx.Hash().String(), nil
}

//CancelTx 取消交易
func (wallet *Wallet) CancelTx(srcTxHash string) (txHash string, err error) {
	client, err := ethclient.DialContext(context.Background(), wallet.rpcURL)
	if err != nil {
		return
	}
	tx, isPending, err := client.TransactionByHash(context.Background(), common.HexToHash(srcTxHash))
	if err != nil {
		return
	}
	if !isPending {
		return "", fmt.Errorf("tx success")
	}
	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		return "", err
	}
	msg, err := tx.AsMessage(types.LatestSignerForChainID(tx.ChainId()), nil)
	if err != nil {
		return
	}
	privateKey, err := wallet.getPrivateKey(msg.From())
	if err != nil {
		return "", err
	}
	gasLimit := uint64(float64(tx.Gas()) * 1.1)
	gasPrice := decimal.NewFromBigInt(tx.GasPrice(), 0).Mul(decimal.NewFromFloat(1.1))
	tx = types.NewTransaction(tx.Nonce(), *tx.To(), big.NewInt(0), gasLimit, gasPrice.BigInt(), nil)
	signer := types.LatestSignerForChainID(chainID)
	signature, err := crypto.Sign(signer.Hash(tx).Bytes(), privateKey)
	if err != nil {
		return "", err
	}
	signedTx, err := tx.WithSignature(signer, signature)
	if err != nil {
		return "", err
	}
	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return "", err
	}
	return signedTx.Hash().String(), nil
}

//Transactions ...
func (wallet *Wallet) Transactions(address string, currency string, page, offset int64, asc bool) (txs string, err error) {
	u, err := url.Parse(fmt.Sprintf("%s/api", txHost))
	if err != nil {
		return
	}
	query := u.Query()
	query.Set("module", "account")
	query.Set("page", fmt.Sprintf("%d", page))
	query.Set("offset", fmt.Sprintf("%d", offset))
	if asc {
		query.Set("sort", "asc")
	} else {
		query.Set("sort", "desc")
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
	defer resp.Body.Close()
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

func (wallet *Wallet) approve(address common.Address, currency string) (allowed decimal.Decimal, err error) {
	uniContract := common.HexToAddress(uniswapContract)
	client, err := ethclient.DialContext(context.Background(), wallet.rpcURL)
	if err != nil {
		return
	}
	defer client.Close()
	chainID, err := client.ChainID(context.Background())
	if err != nil {
		return
	}
	switch currencyType(currency) {
	case USDT:
		contract := common.HexToAddress(contractAddress[USDT])
		usdtToken, err := usdt.NewUsdt(contract, client)
		if err != nil {
			return allowed, err
		}
		albig, err := usdtToken.Allowed(&bind.CallOpts{}, address, uniContract)
		if err != nil {
			return allowed, err
		}
		if albig.Cmp(big.NewInt(0)) > 0 {
			return decimal.NewFromBigInt(albig, 10), nil
		}

		privateKey, err := wallet.getPrivateKey(address)
		if err != nil {
			return allowed, err
		}
		auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
		if err != nil {
			return allowed, err
		}
		tx, err := usdtToken.Approve(auth, uniContract, math.MaxBig256)
		if err != nil {
			return allowed, err
		}
		fmt.Println("txhash:", tx.Hash().String())
		sink := make(chan *usdt.UsdtApproval)
		fmt.Println("watching")
		sub, err := usdtToken.WatchApproval(&bind.WatchOpts{}, sink, []common.Address{address}, []common.Address{uniContract})
		if err != nil {
			return allowed, err
		}
		select {
		case err = <-sub.Err():
			return allowed, err
		case approve := <-sink:
			return decimal.NewFromBigInt(approve.Value, 10), nil
		}
	case EPK:
		contract := common.HexToAddress(contractAddress[EPK])
		epkToken, err := epk.NewEpk(contract, client)
		if err != nil {
			return allowed, err
		}
		albig, err := epkToken.Allowance(&bind.CallOpts{}, address, uniContract)
		if err != nil {
			return allowed, err
		}
		if albig.Cmp(big.NewInt(0)) > 0 {
			return decimal.NewFromBigInt(albig, 10), nil
		}
		privateKey, err := wallet.getPrivateKey(address)
		if err != nil {
			return allowed, err
		}
		auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
		if err != nil {
			return allowed, err
		}
		_, err = epkToken.Approve(auth, uniContract, math.MaxBig256)
		if err != nil {
			return allowed, err
		}
		sink := make(chan *epk.EpkApproval)
		sub, err := epkToken.WatchApproval(&bind.WatchOpts{}, sink, []common.Address{address}, []common.Address{uniContract})
		if err != nil {
			return allowed, err
		}
		select {
		case err = <-sub.Err():
			return allowed, err
		case approve := <-sink:
			return decimal.NewFromBigInt(approve.Value, 10), nil
		}
	case UNI:
		contract := common.HexToAddress(contractAddress[UNI])
		uniToken, err := univ2.NewUniv2(contract, client)
		if err != nil {
			return allowed, err
		}
		albig, err := uniToken.Allowance(&bind.CallOpts{}, address, uniContract)
		if err != nil {
			return allowed, err
		}
		if albig.Cmp(big.NewInt(0)) > 0 {
			return decimal.NewFromBigInt(albig, 10), nil
		}
		privateKey, err := wallet.getPrivateKey(address)
		if err != nil {
			return allowed, err
		}
		auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
		if err != nil {
			return allowed, err
		}
		_, err = uniToken.Approve(auth, uniContract, math.MaxBig256)
		if err != nil {
			return allowed, err
		}
		sink := make(chan *univ2.Univ2Approval)
		sub, err := uniToken.WatchApproval(&bind.WatchOpts{}, sink, []common.Address{address}, []common.Address{uniContract})
		if err != nil {
			return allowed, err
		}
		select {
		case err = <-sub.Err():
			return allowed, err
		case approve := <-sink:
			return decimal.NewFromBigInt(approve.Value, 10), nil
		}
	default:
		return allowed, fmt.Errorf("unsuppoted currency")
	}
}

//UniswapAddLiquidity ...
func (wallet *Wallet) UniswapAddLiquidity(address, tokenA, tokenB, amountADesired, amountBDesired, amountAMin, amountBMin string, deadline string) (txHash string, err error) {
	//converting
	addr := common.HexToAddress(address)
	contact := common.HexToAddress(uniswapContract)
	amAdesiredBig, err := decimal.NewFromString(amountADesired)
	if err != nil {
		return "", fmt.Errorf("amountADesired error")
	}
	amBdesiredBig, err := decimal.NewFromString(amountBDesired)
	if err != nil {
		return "", fmt.Errorf("amountBDesired error")
	}
	amAMinBig, err := decimal.NewFromString(amountAMin)
	if err != nil {
		return "", fmt.Errorf("amountAMin error")
	}
	amBMinBig, err := decimal.NewFromString(amountBMin)
	if err != nil {
		return "", fmt.Errorf("amountBMin error")
	}
	decA, err := getDecimalByCurrency(tokenA)
	if err != nil {
		return "", err
	}
	decB, err := getDecimalByCurrency(tokenB)
	if err != nil {
		return "", err
	}
	amAdesiredBig = amAdesiredBig.Mul(decA)
	amAMinBig = amAMinBig.Mul(decA)
	amBdesiredBig = amBdesiredBig.Mul(decB)
	amBMinBig = amBMinBig.Mul(decB)
	deadlineBig, err := decimal.NewFromString(deadline)
	if err != nil {
		return "", err
	}
	//check approval
	allowed, err := wallet.approve(addr, tokenA)
	if err != nil {
		return "", err
	}

	if allowed.Cmp(amAdesiredBig) < 0 {
		return "", fmt.Errorf("allowance not enougth")
	}
	allowed, err = wallet.approve(addr, tokenB)
	if err != nil {
		return "", err
	}

	if allowed.Cmp(amBdesiredBig) < 0 {
		return "", fmt.Errorf("allowance not enougth")
	}
	//connecting
	client, err := ethclient.DialContext(context.Background(), wallet.rpcURL)
	if err != nil {
		return
	}
	uni, err := uniswap.NewUniswap(contact, client)
	defer client.Close()
	chainID, err := client.ChainID(context.Background())
	if err != nil {
		return
	}
	privateKey, err := wallet.getPrivateKey(addr)
	if err != nil {
		return "", err
	}
	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		return "", err
	}
	tx, err := uni.AddLiquidity(auth, common.HexToAddress(contractAddress[currencyType(tokenA)]), common.HexToAddress(contractAddress[currencyType(tokenB)]), amAdesiredBig.BigInt(), amBdesiredBig.BigInt(), amAMinBig.BigInt(), amBMinBig.BigInt(), addr, deadlineBig.BigInt())
	if err != nil {
		return "", err
	}
	txHash = tx.Hash().String()
	return
}

//UniswapRemoveLiquidity ...
func (wallet *Wallet) UniswapRemoveLiquidity(address, tokenA, tokenB, liquidity, amountAMin, amountBMin, deadline string) (txHash string, err error) {
	//converting
	addr := common.HexToAddress(address)
	contract := common.HexToAddress(uniswapContract)
	liquidityBig, err := decimal.NewFromString(liquidity)
	if err != nil {
		return "", fmt.Errorf("liquidity error")
	}
	amAMinBig, err := decimal.NewFromString(amountAMin)
	if err != nil {
		return "", fmt.Errorf("amountAMin error")
	}
	amBMinBig, err := decimal.NewFromString(amountBMin)
	if err != nil {
		return "", fmt.Errorf("amountBMin error")
	}
	decA, err := getDecimalByCurrency(tokenA)
	if err != nil {
		return "", err
	}
	decB, err := getDecimalByCurrency(tokenB)
	if err != nil {
		return "", err
	}
	decUni, err := getDecimalByCurrency("UNI")
	if err != nil {
		return "", err
	}
	amAMinBig = amAMinBig.Mul(decA)
	liquidityBig = liquidityBig.Mul(decUni)
	amBMinBig = amBMinBig.Mul(decB)
	deadlineBig, err := decimal.NewFromString(deadline)
	if err != nil {
		return "", err
	}
	//check approval
	allowed, err := wallet.approve(addr, "UNI")
	if err != nil {
		return "", err
	}

	if allowed.Cmp(liquidityBig) < 0 {
		return "", fmt.Errorf("allowance not enougth")
	}
	//connecting
	client, err := ethclient.DialContext(context.Background(), wallet.rpcURL)
	if err != nil {
		return
	}
	uni, err := uniswap.NewUniswap(contract, client)
	defer client.Close()
	chainID, err := client.ChainID(context.Background())
	if err != nil {
		return
	}
	privateKey, err := wallet.getPrivateKey(addr)
	if err != nil {
		return "", err
	}
	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		return "", err
	}
	tx, err := uni.RemoveLiquidity(auth, common.HexToAddress(contractAddress[currencyType(tokenA)]), common.HexToAddress(contractAddress[currencyType(tokenB)]), liquidityBig.BigInt(), amAMinBig.BigInt(), amBMinBig.BigInt(), addr, deadlineBig.BigInt())
	if err != nil {
		return "", err
	}
	txHash = tx.Hash().String()
	return
}

//UniswapExactTokenForTokens ...
func (wallet *Wallet) UniswapExactTokenForTokens(address, tokenA, tokenB, amountIn, amountOutMin, deadline string) (txHash string, err error) {

	//converting
	addr := common.HexToAddress(address)
	contract := common.HexToAddress(uniswapContract)
	path := []common.Address{}
	amInBig, err := decimal.NewFromString(amountIn)
	if err != nil {
		return "", fmt.Errorf("amountIn error")
	}
	amOutMinBig, err := decimal.NewFromString(amountOutMin)
	if err != nil {
		return "", fmt.Errorf("amountIn error")
	}
	decA, err := getDecimalByCurrency(tokenA)
	if err != nil {
		return "", err
	}
	decB, err := getDecimalByCurrency(tokenB)
	if err != nil {
		return "", err
	}
	path = append(path, common.HexToAddress(contractAddress[currencyType(tokenA)]))
	amInBig = amInBig.Mul(decA)
	path = append(path, common.HexToAddress(contractAddress[currencyType(tokenB)]))
	amOutMinBig = amOutMinBig.Mul(decB)
	deadlineInt, err := decimal.NewFromString(deadline)
	if err != nil {
		return "", err
	}
	//check approval
	allowed, err := wallet.approve(addr, tokenA)
	if err != nil {
		return "", err
	}

	if allowed.Cmp(amInBig) < 0 {
		return "", fmt.Errorf("allowance not enougth")
	}
	//connecting
	client, err := ethclient.DialContext(context.Background(), wallet.rpcURL)
	if err != nil {
		return
	}
	uni, err := uniswap.NewUniswap(contract, client)
	defer client.Close()
	chainID, err := client.ChainID(context.Background())
	if err != nil {
		return
	}
	privateKey, err := wallet.getPrivateKey(addr)
	if err != nil {
		return "", err
	}
	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		return "", err
	}
	tx, err := uni.SwapExactTokensForTokens(auth, amInBig.BigInt(), amOutMinBig.BigInt(), path, addr, deadlineInt.BigInt())
	if err != nil {
		return "", err
	}
	txHash = tx.Hash().String()
	return
}

//Amounts ...
type Amounts struct {
	AmountIn  string
	AmountOut string
}

//UniswapGetAmountsOut ...
func (wallet *Wallet) UniswapGetAmountsOut(tokenA, tokenB, amountIn string) (amounts *Amounts, err error) {
	client, err := ethclient.DialContext(context.Background(), wallet.rpcURL)
	if err != nil {
		return
	}
	defer client.Close()
	contact := common.HexToAddress(uniswapContract)
	uni, err := uniswap.NewUniswap(contact, client)
	path := []common.Address{}
	amInBig, err := decimal.NewFromString(amountIn)
	if err != nil {
		return nil, fmt.Errorf("amountIn error")
	}
	decA := decimal.Zero
	decB := decimal.Zero
	switch currencyType(tokenA) {
	case USDT:
		decA, _ = getDecimalByCurrency("USDT")
	case EPK:
		decA, _ = getDecimalByCurrency("EPK")
	default:
		return nil, fmt.Errorf("unsuppoted currency")
	}
	path = append(path, common.HexToAddress(contractAddress[currencyType(tokenA)]))
	amInBig = amInBig.Mul(decA)
	switch currencyType(tokenB) {
	case USDT:
		decB, _ = getDecimalByCurrency("USDT")
	case EPK:
		decB, _ = getDecimalByCurrency("EPK")
	default:
		return nil, fmt.Errorf("unsuppoted currency")
	}
	path = append(path, common.HexToAddress(contractAddress[currencyType(tokenB)]))
	amts, err := uni.GetAmountsOut(&bind.CallOpts{}, amInBig.BigInt(), path)
	if err != nil {
		return nil, err
	}
	if len(amts) != 2 {
		return nil, fmt.Errorf("amounts error")
	}
	amounts = &Amounts{}
	amounts.AmountIn = decimal.NewFromBigInt(amts[0], 0).Div(decA).String()
	amounts.AmountOut = decimal.NewFromBigInt(amts[1], 0).Div(decB).String()
	return
}

//UniswapInfo ...
type UniswapInfo struct {
	EPK           string
	USDT          string
	UNI           string
	Share         string
	LastBlockTime int64
}

//UniswapInfo ...
func (wallet *Wallet) UniswapInfo(address string) (info *UniswapInfo, err error) {
	client, err := ethclient.DialContext(context.Background(), wallet.rpcURL)
	if err != nil {
		return
	}
	defer client.Close()
	userAddress := common.HexToAddress(address)
	contract := common.HexToAddress(contractAddress[UNI])
	uni, err := univ2.NewUniv2(contract, client)
	if err != nil {
		return
	}
	info = &UniswapInfo{}
	totalSupply, err := uni.TotalSupply(&bind.CallOpts{})
	if err != nil {
		return
	}
	userBalance, err := uni.BalanceOf(&bind.CallOpts{}, userAddress)
	if err != nil {
		return
	}
	info.Share = decimal.NewFromBigInt(userBalance, 0).Div(decimal.NewFromBigInt(totalSupply, 0)).String()
	token0, err := uni.Token0(&bind.CallOpts{})
	token1, err := uni.Token1(&bind.CallOpts{})
	reserves, err := uni.GetReserves(&bind.CallOpts{})
	decUSDT, _ := getDecimalByCurrency("USDT")
	decEPK, _ := getDecimalByCurrency("EPK")
	decUNI, _ := getDecimalByCurrency("UNI")
	fmt.Println(token0.String())
	if token0.String() == contractAddress[USDT] {
		info.USDT = decimal.NewFromBigInt(reserves.Reserve0, 0).Div(decUSDT).String()
	} else if token0.String() == contractAddress[EPK] {
		info.EPK = decimal.NewFromBigInt(reserves.Reserve0, 0).Div(decEPK).String()
	}
	if token1.String() == contractAddress[USDT] {
		info.USDT = decimal.NewFromBigInt(reserves.Reserve1, 0).Div(decUSDT).String()
	} else if token1.String() == contractAddress[EPK] {
		info.EPK = decimal.NewFromBigInt(reserves.Reserve1, 0).Div(decEPK).String()
	}
	info.UNI = decimal.NewFromBigInt(userBalance, 0).Div(decUNI).String()
	fmt.Println(userBalance, decimal.NewFromBigInt(userBalance, 0).Div(decUNI), info.UNI)
	info.LastBlockTime = int64(reserves.BlockTimestampLast)
	return
}

func (wallet *Wallet) getPrivateKey(address common.Address) (priKey *ecdsa.PrivateKey, err error) {
	var account accounts.Account
	find := false
	for _, acc := range wallet.hdWallet.Accounts() {
		if acc.Address == address {
			find = true
			account = acc
			break
		}
	}
	if !find {
		return nil, fmt.Errorf("account not found")
	}
	return wallet.hdWallet.PrivateKey(account)

}

func getDecimalByCurrency(currency string) (dec decimal.Decimal, err error) {
	switch currencyType(currency) {
	case USDT:
		return decimal.NewFromString("1000000")
		// return decimal.NewFromString("1000000000000000000")
	case EPK:
		return decimal.NewFromString("1000000000000000000")
	case UNI:
		return decimal.NewFromString("1000000000000000000")
	default:
		return dec, fmt.Errorf("not suppoted")
	}
}
