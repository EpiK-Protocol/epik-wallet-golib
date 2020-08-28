package epik

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/EpiK-Protocol/go-epik/api/client"
	"github.com/EpiK-Protocol/go-epik/chain/types"
	epikwallet "github.com/EpiK-Protocol/go-epik/chain/wallet"
	"github.com/ipfs/go-cid"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/specs-actors/actors/abi"
	"github.com/filecoin-project/specs-actors/actors/crypto"
)

//EpikWallet wallet
type EpikWallet struct {
	epikWallet *epikwallet.Wallet
	rpcURL     string
	header     http.Header
	ctx        context.Context
}

//PrivateKey ...
type PrivateKey struct {
	KeyType    string
	PrivateKey string
}

//EPKMessage ...
type EPKMessage struct {
	Version  int64
	From     string
	To       string
	Value    string
	Nonce    int64
	GasLimit int64
	GasPrice string
	Method   string
	Params   []byte
}

//NewWallet ...
func NewWallet() (w *EpikWallet, err error) {
	ks := epikwallet.NewMemKeyStore()
	wa, err := epikwallet.NewWallet(ks)
	if err != nil {
		return nil, err
	}
	w = &EpikWallet{
		epikWallet: wa,
	}
	return w, nil
}

//GenerateKey t:bls,secp256k1
func (w *EpikWallet) GenerateKey(t string, seed []byte) (addrStr string, err error) {
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
func (w *EpikWallet) AddrList() (addrs []string, err error) {
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
func (w *EpikWallet) HasAddr(addr string) (has bool, err error) {
	ad, err := address.NewFromString(addr)
	if err != nil {
		return false, err
	}
	return w.epikWallet.HasKey(ad)
}

//Export ...
func (w *EpikWallet) Export(addr string) (privateKey PrivateKey, err error) {
	privateKey = PrivateKey{}
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
	privateKey.PrivateKey = base64.StdEncoding.EncodeToString(keyInfo.PrivateKey)
	return
}

//Import ...
func (w *EpikWallet) Import(privateKey PrivateKey) (addr string, err error) {
	switch strings.ToLower(privateKey.KeyType) {
	case "bks", "secp256k1":
	default:
		return "", fmt.Errorf("Key Type (%s) Not Suppoted", privateKey.KeyType)

	}
	pk, err := base64.StdEncoding.DecodeString(privateKey.PrivateKey)
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
func (w *EpikWallet) SetDefault(addr string) (err error) {
	ad, err := address.NewFromString(addr)
	if err != nil {
		return err
	}
	return w.epikWallet.SetDefault(ad)
}

//SetRPC ...
func (w *EpikWallet) SetRPC(url string, token string) (err error) {
	w.rpcURL = url
	w.header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	return
}

//Balance ...
func (w *EpikWallet) Balance(addr string) (balance string, err error) {
	ad, err := address.NewFromString(addr)
	fullAPI, _, err := client.NewFullNodeRPC(w.rpcURL, w.header)
	bal, err := fullAPI.WalletBalance(w.ctx, ad)
	balance = bal.String()
	return
}

//Send ...
func (w *EpikWallet) Send(to string, amount string) (cid cid.Cid, err error) {
	fromAddr, err := w.epikWallet.GetDefault()
	if err != nil {
		return cid, err
	}
	toAddr, err := address.NewFromString(to)
	fullAPI, _, err := client.NewFullNodeRPC(w.rpcURL, w.header)
	if err != nil {
		return
	}
	head, err := fullAPI.ChainHead(w.ctx)
	gasPrice, err := fullAPI.MpoolEstimateGasPrice(w.ctx, 10, fromAddr, 10000, head.Key())
	epk, err := types.ParseFIL(amount)
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
	cid, err = fullAPI.MpoolPush(context.Background(), signedMsg)
	return
}

//MessageList ...
func (w *EpikWallet) MessageList(toHeight int64, addr string) (messages []*EPKMessage, err error) {
	ad, err := address.NewFromString(addr)
	if err != nil {
		return
	}
	fullAPI, _, err := client.NewFullNodeRPC(w.rpcURL, w.header)
	if err != nil {
		return
	}
	head, err := fullAPI.ChainHead(w.ctx)
	if err != nil {
		return
	}
	cids, err := fullAPI.StateListMessages(w.ctx, &types.Message{From: ad, To: ad}, head.Key(), abi.ChainEpoch(toHeight))
	if err != nil {
		return
	}
	for _, cid := range cids {
		message, err := fullAPI.ChainGetMessage(w.ctx, cid)
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
		messages = append(messages, epkmsg)
	}
	return
}
