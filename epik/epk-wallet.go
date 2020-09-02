package epik

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/EpiK-Protocol/go-epik/api/client"
	"github.com/EpiK-Protocol/go-epik/chain/types"
	epikwallet "github.com/EpiK-Protocol/go-epik/chain/wallet"

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
func (w *Wallet) GenerateKey(t string, seed []byte) (addrStr string, err error) {
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
func (w *Wallet) Export(addr string) (privateKey *PrivateKey, err error) {
	privateKey = &PrivateKey{}
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
	privateKey.KeyType = keyInfo.Type
	privateKey.PrivateKey = hex.EncodeToString(keyInfo.PrivateKey)
	return
}

//Import ...
func (w *Wallet) Import(privateKey *PrivateKey) (addr string, err error) {
	switch strings.ToLower(privateKey.KeyType) {
	case "bks", "secp256k1":
	default:
		return "", fmt.Errorf("Key Type (%s) Not Suppoted", privateKey.KeyType)

	}
	pk, err := hex.DecodeString(privateKey.PrivateKey)
	if err != nil {
		return "", err
	}
	keyInfo := &types.KeyInfo{
		Type:       privateKey.KeyType,
		PrivateKey: pk,
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
	bal, err := fullAPI.WalletBalance(context.Background(), ad)
	balance = bal.String()
	return
}

//Send ...
func (w *Wallet) Send(to string, amount string) (cidStr string, err error) {
	fromAddr, err := w.epikWallet.GetDefault()
	if err != nil {
		return "", err
	}
	toAddr, err := address.NewFromString(to)
	fullAPI, _, err := client.NewFullNodeRPC(w.rpcURL, w.header)
	if err != nil {
		return
	}
	head, err := fullAPI.ChainHead(context.Background())
	gasPrice, err := fullAPI.MpoolEstimateGasPrice(context.Background(), 10, fromAddr, 10000, head.Key())
	epk, err := types.ParseEPK(amount)
	msg := types.Message{
		From:     fromAddr,
		To:       toAddr,
		Value:    types.BigInt(epk),
		GasPrice: gasPrice,
		GasLimit: 10000,
	}
	signature, err := w.epikWallet.Sign(context.Background(), fromAddr, msg.Cid().Bytes())
	if err != nil {
		fmt.Println(err)
	}
	signedMsg := &types.SignedMessage{
		Message:   msg,
		Signature: *signature,
	}
	fmt.Println(signedMsg)
	c, err := fullAPI.MpoolPush(context.Background(), signedMsg)
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
	cids, err := fullAPI.StateListMessages(context.Background(), &types.Message{From: ad, To: ad}, head.Key(), abi.ChainEpoch(toHeight))
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
	messages = string(bs)
	return
}
