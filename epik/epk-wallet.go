package epik

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/EpiK-Protocol/go-epik/api/client"
	"github.com/EpiK-Protocol/go-epik/chain/types"
	epikwallet "github.com/EpiK-Protocol/go-epik/chain/wallet"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil/hdkeychain"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/shopspring/decimal"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/specs-actors/actors/abi"
	"github.com/filecoin-project/specs-actors/actors/crypto"
)

//Wallet wallet
type Wallet struct {
	epikWallet *epikwallet.Wallet
	rpcURL     string
	header     http.Header
}

//PrivateKey ...
type PrivateKey struct {
	KeyType    string
	PrivateKey string
}

//EPKMessage ...
type EPKMessage struct {
	Version  int64  `json:"version"`
	From     string `json:"from"`
	To       string `json:"to"`
	Value    string `json:"value"`
	Nonce    int64  `json:"nonce"`
	GasLimit int64  `json:"gas_limit"`
	GasPrice string `json:"gas_price"`
	Method   string `json:"method"`
	Params   []byte `json:"params"`
}

//NewWallet ...
func NewWallet() (w *Wallet, err error) {
	ks := epikwallet.NewMemKeyStore()
	wa, err := epikwallet.NewWallet(ks)
	if err != nil {
		return nil, err
	}
	w = &Wallet{
		epikWallet: wa,
	}
	return w, nil
}

//GenerateKey t:bls,secp256k1
func (w *Wallet) GenerateKey(t string, seed []byte, path string) (addrStr string, err error) {
	seed, err = epikHDPathSeed(seed, path)
	if err != nil {
		return "", err
	}
	fmt.Println(seed)
	var addr address.Address
	switch strings.ToLower(t) {
	case "bls":
		addr, err = w.epikWallet.GenerateKeyFromSeed(crypto.SigTypeBLS, seed)
	case "secp256k1":
		addr, err = w.epikWallet.GenerateKeyFromSeed(crypto.SigTypeSecp256k1, seed)
	default:
		return "", fmt.Errorf("SigType not suppot")
	}
	if err != nil {
		return "", err
	}
	return addr.String(), nil
}

func epikHDPathSeed(seed []byte, path string) (pathSeed []byte, err error) {
	masterKey, err := hdkeychain.NewMaster(seed, &chaincfg.MainNetParams)
	if err != nil {
		return nil, err
	}
	p, err := accounts.ParseDerivationPath(path)
	if err != nil {
		return nil, err
	}
	key := masterKey
	for _, n := range p {
		key, err = key.Child(n)
		if err != nil {
			return nil, err
		}
	}
	rawSeed := reflect.ValueOf(key).Elem().FieldByName("key").Bytes()
	seed = make([]byte, 32)
	if len(rawSeed) <= 32 {
		copy(seed[32-len(rawSeed):], rawSeed[:])
	} else {
		copy(seed[:], rawSeed[:])
	}

	return seed, nil

}

//AddrList ...
func (w *Wallet) AddrList() (addrs []string, err error) {
	ads, err := w.epikWallet.ListAddrs()
	if err != nil {
		return nil, err
	}
	for _, ad := range ads {
		addrs = append(addrs, ad.String())
	}
	return
}

//HasAddr ...
func (w *Wallet) HasAddr(addr string) (has bool) {
	ad, err := address.NewFromString(addr)
	if err != nil {
		return false
	}
	has, _ = w.epikWallet.HasKey(ad)
	return
}

//Export ...
func (w *Wallet) Export(addr string) (privateKey string, err error) {
	ad, err := address.NewFromString(addr)
	if err != nil {
		return
	}
	has, err := w.epikWallet.HasKey(ad)
	if err != nil {
		return
	}
	if !has {
		return privateKey, fmt.Errorf("addr not found")
	}
	keyInfo, err := w.epikWallet.Export(ad)
	if err != nil {
		return
	}
	data, err := json.Marshal(keyInfo)
	if err != nil {
		return "", err
	}
	privateKey = hex.EncodeToString(data)
	return
}

//Import ...
func (w *Wallet) Import(privateKey string) (addr string, err error) {

	data, err := hex.DecodeString(privateKey)
	if err != nil {
		return "", err
	}
	keyInfo := &types.KeyInfo{}
	err = json.Unmarshal(data, keyInfo)
	if err != nil {
		return "", err
	}
	ad, err := w.epikWallet.Import(keyInfo)
	if err != nil {
		return "", err
	}
	return ad.String(), nil
}

//SetDefault ...
func (w *Wallet) SetDefault(addr string) (err error) {
	ad, err := address.NewFromString(addr)
	if err != nil {
		return err
	}
	return w.epikWallet.SetDefault(ad)
}

//Sign ...
func (w *Wallet) Sign(addr string, hash []byte) (signature []byte, err error) {
	ad, err := w.epikWallet.GetDefault()
	if addr != "" {
		ad, err = address.NewFromString(addr)
	}
	if err != nil {
		return
	}
	sign, err := w.epikWallet.Sign(context.Background(), ad, hash)
	if err != nil {
		return
	}
	return sign.MarshalBinary()
}

//SetRPC ...
func (w *Wallet) SetRPC(url string, token string) (err error) {
	w.rpcURL = url
	w.header = http.Header{}
	w.header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	return
}

//Balance ...
func (w *Wallet) Balance(addr string) (balance string, err error) {
	ad, err := address.NewFromString(addr)
	fullAPI, _, err := client.NewFullNodeRPC(w.rpcURL, w.header)
	if err != nil {
		return "", err
	}
	bal, err := fullAPI.WalletBalance(context.Background(), ad)
	if err != nil {
		return "", err
	}
	balance = BigIntDiv(bal.String(), 18)
	return
}

//Send ...
func (w *Wallet) Send(to string, amount string) (cidStr string, err error) {
	fromAddr, err := w.epikWallet.GetDefault()
	if err != nil {
		return "", err
	}
	toAddr, err := address.NewFromString(to)
	if err != nil {
		return
	}
	fullAPI, _, err := client.NewFullNodeRPC(w.rpcURL, w.header)
	if err != nil {
		return
	}
	head, err := fullAPI.ChainHead(context.Background())
	if err != nil {
		return
	}
	gasPrice, err := fullAPI.MpoolEstimateGasPrice(context.Background(), 10, fromAddr, 10000, head.Key())
	if err != nil {
		return
	}
	nonce, err := fullAPI.MpoolGetNonce(context.Background(), fromAddr)
	if err != nil {
		return
	}
	epk, err := types.ParseEPK(amount)
	if err != nil {
		return
	}
	msg := types.Message{
		From:     fromAddr,
		To:       toAddr,
		Value:    types.BigInt(epk),
		GasPrice: gasPrice,
		GasLimit: 10000,
		Nonce:    nonce,
	}
	signature, err := w.epikWallet.Sign(context.Background(), fromAddr, msg.Cid().Bytes())
	if err != nil {
		return "", err
	}
	signedMsg := &types.SignedMessage{
		Message:   msg,
		Signature: *signature,
	}
	fmt.Println(signedMsg)
	c, err := fullAPI.MpoolPush(context.Background(), signedMsg)
	if err != nil {
		return "", err
	}
	return c.String(), nil
}

//MessageList ...
func (w *Wallet) MessageList(toHeight int64, addr string) (messages string, err error) {
	ad, err := address.NewFromString(addr)
	if err != nil {
		return
	}
	fullAPI, _, err := client.NewFullNodeRPC(w.rpcURL, w.header)
	if err != nil {
		return
	}
	head, err := fullAPI.ChainHead(context.Background())
	if err != nil {
		return
	}
	from, err := fullAPI.StateListMessages(context.Background(), &types.Message{From: ad}, head.Key(), abi.ChainEpoch(toHeight))
	to, err := fullAPI.StateListMessages(context.Background(), &types.Message{To: ad}, head.Key(), abi.ChainEpoch(toHeight))
	cids := append(from, to...)
	if err != nil {
		return
	}
	ms := []*EPKMessage{}
	for _, cid := range cids {
		message, err := fullAPI.ChainGetMessage(context.Background(), cid)
		if err != nil {
			continue
		}
		epkmsg := &EPKMessage{
			Version:  message.Version,
			From:     message.From.String(),
			To:       message.To.String(),
			Value:    message.Value.String(),
			Nonce:    int64(message.Nonce),
			GasLimit: message.GasLimit,
			GasPrice: message.GasPrice.String(),
			Method:   message.Method.String(),
			Params:   message.Params,
		}
		ms = append(ms, epkmsg)
	}
	bs, err := json.Marshal(&ms)
	if err != nil {
		return "", err
	}
	messages = string(bs)
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
