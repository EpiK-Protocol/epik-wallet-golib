package epik

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/EpiK-Protocol/epik-wallet-golib/epik/client"
	"github.com/EpiK-Protocol/epik-wallet-golib/epik/wallet"
	"github.com/EpiK-Protocol/go-epik/api"
	"github.com/EpiK-Protocol/go-epik/chain/actors"
	"github.com/EpiK-Protocol/go-epik/chain/actors/builtin/expert"
	"github.com/EpiK-Protocol/go-epik/chain/actors/builtin/miner"
	"github.com/EpiK-Protocol/go-epik/chain/actors/builtin/retrieval"
	vesting2 "github.com/EpiK-Protocol/go-epik/chain/actors/builtin/vesting"
	"github.com/EpiK-Protocol/go-epik/chain/types"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil/hdkeychain"
	"github.com/ethereum/go-ethereum/accounts"
	fminer "github.com/filecoin-project/specs-actors/v2/actors/builtin/miner"
	"github.com/ipfs/go-cid"
	"github.com/shopspring/decimal"
	"golang.org/x/xerrors"

	"github.com/filecoin-project/go-address"
	jsonrpc "github.com/filecoin-project/go-jsonrpc"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/go-state-types/crypto"
	"github.com/filecoin-project/specs-actors/v2/actors/builtin"
	"github.com/filecoin-project/specs-actors/v2/actors/builtin/expertfund"
	"github.com/filecoin-project/specs-actors/v2/actors/builtin/vesting"
	"github.com/filecoin-project/specs-actors/v2/actors/builtin/vote"
)

//Wallet wallet
type Wallet struct {
	epikWallet *wallet.LocalWallet
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
	ks := wallet.NewMemKeyStore()
	wa, err := wallet.NewWallet(ks)
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
	var addr address.Address
	switch strings.ToLower(t) {
	case "bls":
		addr, err = w.epikWallet.WalletNewFromSeed(types.KTBLS, seed)
	case "secp256k1":
		addr, err = w.epikWallet.WalletNewFromSeed(types.KTSecp256k1, seed)
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
		key, err = key.Derive(n)
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
	ads, err := w.epikWallet.WalletList(context.Background())
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
	has, _ = w.epikWallet.WalletHas(context.Background(), ad)
	return
}

//Export ...
func (w *Wallet) Export(addr string) (privateKey string, err error) {
	ad, err := address.NewFromString(addr)
	if err != nil {
		return
	}
	has, err := w.epikWallet.WalletHas(context.Background(), ad)
	if err != nil {
		return
	}
	if !has {
		return privateKey, fmt.Errorf("addr not found")
	}
	keyInfo, err := w.epikWallet.WalletExport(context.Background(), ad)
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
	ad, err := w.epikWallet.WalletImport(context.Background(), keyInfo)
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
	return w.epikWallet.SetDefault(context.Background(), ad)
}

//Sign ...
func (w *Wallet) Sign(addr string, hash []byte) (signature []byte, err error) {
	ad, err := w.epikWallet.GetDefault(context.Background())
	if addr != "" {
		ad, err = address.NewFromString(addr)
	}
	if err != nil {
		return
	}
	sign, err := w.epikWallet.WalletSign(context.Background(), ad, hash)
	if err != nil {
		return
	}
	return sign.MarshalBinary()
}

func (w *Wallet) SignAndSendMessage(addr string, message string) (cidStr string, err error) {
	msg := &types.Message{}
	err = json.Unmarshal([]byte(message), msg)
	if err != nil {
		return
	}
	fromAddr, err := w.epikWallet.GetDefault(context.Background())
	if err != nil {
		return "", err
	}
	fullAPI, closer, err := client.NewFullNodeRPC(context.Background(), w.rpcURL, w.header)
	if err != nil {
		return
	}
	msg.From = fromAddr
	defer closer()
	c, err := w.sendMessage(fullAPI, msg)
	if err != nil {
		return "", err
	}
	return c.String(), nil
}

func (w *Wallet) SignCID(addr string, cidStr string) (signature []byte, err error) {
	ad, err := w.epikWallet.GetDefault(context.Background())
	if addr != "" {
		ad, err = address.NewFromString(addr)
	}
	if err != nil {
		return
	}
	cID, err := cid.Decode(cidStr)
	if err != nil {
		return
	}
	s, err := w.epikWallet.WalletSign(context.Background(), ad, cID.Bytes())
	if err != nil {
		return
	}
	return s.MarshalBinary()
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
	fullAPI, closer, err := client.NewFullNodeRPC(context.Background(), w.rpcURL, w.header)
	if err != nil {
		return "", err
	}
	defer closer()
	bal, err := fullAPI.WalletBalance(context.Background(), ad)
	if err != nil {
		return "", err
	}
	balance = decimal.NewFromBigInt(bal.Int, -18).String()
	return
}

//Send ...
func (w *Wallet) Send(to string, amount string) (cidStr string, err error) {
	fromAddr, err := w.epikWallet.GetDefault(context.Background())
	if err != nil {
		return "", err
	}
	toAddr, err := address.NewFromString(to)
	if err != nil {
		return
	}
	fullAPI, closer, err := client.NewFullNodeRPC(context.Background(), w.rpcURL, w.header)
	if err != nil {
		return
	}
	defer closer()
	// gasPrice, err := fullAPI.MpoolEstimateGasPrice(context.Background(), 10, fromAddr, 10000, head.Key())
	// if err != nil {
	// 	return
	// }

	epk, err := types.ParseEPK(amount)
	if err != nil {
		return
	}
	msg := &types.Message{
		From:  fromAddr,
		To:    toAddr,
		Value: types.BigInt(epk),
	}
	c, err := w.sendMessage(fullAPI, msg)
	if err != nil {
		return "", err
	}
	return c.String(), nil
}

func (w *Wallet) CreateSendMessage(to string, amount string) (message string, err error) {
	fromAddr, err := w.epikWallet.GetDefault(context.Background())
	if err != nil {
		return "", err
	}
	toAddr, err := address.NewFromString(to)
	if err != nil {
		return
	}

	epk, err := types.ParseEPK(amount)
	if err != nil {
		return
	}
	msg := &types.Message{
		From:  fromAddr,
		To:    toAddr,
		Value: types.BigInt(epk),
	}
	data, _ := json.Marshal(msg)
	message = string(data)
	return
}

func (w *Wallet) GasEstimateGasLimit(actor string) (gasLimit string, err error) {
	from, err := w.epikWallet.GetDefault(context.Background())
	if err != nil {
		return
	}
	msg := &types.Message{}
	switch actor {
	case "transfer", "":
		msg.From = from
		msg.To = from
	default:
		return "0", fmt.Errorf("actor not found")
	}
	node, closer, err := client.NewFullNodeRPC(context.Background(), w.rpcURL, w.header)
	if err != nil {
		return
	}
	defer closer()
	limit, err := node.GasEstimateGasLimit(context.Background(), msg, types.EmptyTSK)
	if err != nil {
		return
	}
	cap, err := node.GasEstimateFeeCap(context.Background(), msg, 20, types.EmptyTSK)
	if err != nil {
		return
	}
	return big.Mul(cap, big.NewInt(limit)).String(), nil
}

func (w *Wallet) MessageCID(message string) (cidStr string, err error) {
	msg := types.Message{}
	err = json.Unmarshal([]byte(message), &msg)
	if err != nil {
		return
	}
	cidStr = msg.Cid().String()
	return
}

func (w *Wallet) SendRawMessage(message string, signature []byte) (cidStr string, err error) {
	msg := types.Message{}
	err = json.Unmarshal([]byte(message), &msg)
	if err != nil {
		return
	}
	sign := &crypto.Signature{}
	err = sign.UnmarshalBinary(signature)
	if err != nil {
		return
	}
	node, closer, err := client.NewFullNodeRPC(context.Background(), w.rpcURL, w.header)
	if err != nil {
		return
	}
	defer closer()

	msg.Nonce, err = node.MpoolGetNonce(context.Background(), msg.From)
	if err != nil {
		return
	}
	msg.GasFeeCap, err = node.GasEstimateFeeCap(context.Background(), &msg, 0, types.EmptyTSK)
	if err != nil {
		return
	}
	msg.GasLimit, err = node.GasEstimateGasLimit(context.Background(), &msg, types.EmptyTSK)
	if err != nil {
		return
	}
	signedMsg := &types.SignedMessage{
		Message:   msg,
		Signature: *sign,
	}
	fmt.Println(json.Marshal(signedMsg))
	c, err := node.MpoolPush(context.Background(), signedMsg)
	if err != nil {
		return
	}
	return c.String(), nil
}

//MessageReceipt ...
func (w *Wallet) MessageReceipt(cidStr string) (status string, err error) {
	cidHash, err := cid.Parse(cidStr)
	fullAPI, closer, err := client.NewFullNodeRPC(context.Background(), w.rpcURL, w.header)
	if err != nil {
		return
	}
	defer closer()
	msg, err := fullAPI.ChainGetMessage(context.Background(), cidHash)
	if err != nil {
		return "", err
	}
	if msg == nil {
		return "", fmt.Errorf("message not found")
	}
	receipt, err := fullAPI.StateGetReceipt(context.Background(), cidHash, types.EmptyTSK)
	if err != nil {
		return "", err
	}
	if receipt == nil {
		return "pending", nil
	}
	if receipt.ExitCode.IsSuccess() {
		return "success", nil
	} else if receipt.ExitCode.IsSendFailure() {
		return "failed", nil
	} else if receipt.ExitCode.IsError() {
		return "error", fmt.Errorf(receipt.ExitCode.Error())
	}
	return "", fmt.Errorf("not suppoted exitCode")
}

func (w *Wallet) sendMessage(fullAPI api.FullNode, msg *types.Message) (cidStr cid.Cid, err error) {
	msg.Nonce, err = fullAPI.MpoolGetNonce(context.Background(), msg.From)
	if err != nil {
		return
	}

	msg.GasFeeCap, err = fullAPI.GasEstimateFeeCap(context.Background(), msg, 20, types.EmptyTSK)
	if err != nil {
		return
	}
	msg.GasLimit, err = fullAPI.GasEstimateGasLimit(context.Background(), msg, types.EmptyTSK)
	if err != nil {
		return
	}
	msg.GasLimit = int64(float64(msg.GasLimit) * 1.25)
	signature, err := w.epikWallet.WalletSign(context.Background(), msg.From, msg.Cid().Bytes())
	if err != nil {
		return cid.Undef, err
	}
	signedMsg := &types.SignedMessage{
		Message:   *msg,
		Signature: *signature,
	}
	return fullAPI.MpoolPush(context.Background(), signedMsg)
}

func (w *Wallet) CoinbaseInfo(addr string) (infoJSON string, err error) {
	ctx := context.Background()
	from := address.Address{}
	if addr != "" {
		from, err = address.NewFromString(addr)
		if err != nil {
			return
		}
	} else {
		from, err = w.epikWallet.GetDefault(ctx)
		if err != nil {
			return
		}
	}
	if from.Empty() {
		return "", fmt.Errorf("no address")
	}
	node, closer, err := client.NewFullNodeRPC(ctx, w.rpcURL, w.header)
	if err != nil {
		return
	}
	defer closer()
	fromID, _ := address.NewIDAddress(0)
	if from.Protocol() == address.ID {
		fromID = from
	} else {
		fromID, err = node.StateLookupID(ctx, from, types.EmptyTSK)
		if err != nil {
			return
		}
	}
	info, err := node.StateCoinbase(ctx, fromID, types.EmptyTSK)
	if err != nil {
		return
	}
	result := &struct {
		Total    decimal.Decimal
		Vesting  decimal.Decimal
		Vested   decimal.Decimal
		Coinbase address.Address
	}{
		Total:    decimal.NewFromBigInt(info.Total.Int, -18),
		Vesting:  decimal.NewFromBigInt(info.Vesting.Int, -18),
		Vested:   decimal.NewFromBigInt(info.Vested.Int, -18),
		Coinbase: fromID,
	}
	data, _ := json.Marshal(result)
	return string(data), nil
}

func (w *Wallet) CoinbaseWithdraw() (cidStr string, err error) {
	from, err := w.epikWallet.GetDefault(context.Background())
	if err != nil {
		return
	}

	node, closer, err := client.NewFullNodeRPC(context.Background(), w.rpcURL, w.header)
	if err != nil {
		return
	}
	defer closer()
	fromID, _ := address.NewIDAddress(0)
	if from.Protocol() == address.ID {
		fromID = from
	} else {
		fromID, err = node.StateLookupID(context.Background(), from, types.EmptyTSK)
		if err != nil {
			return
		}
	}
	info, err := node.StateCoinbase(context.Background(), fromID, types.EmptyTSK)
	if err != nil {
		return
	}
	params, err := actors.SerializeParams(&vesting.WithdrawBalanceParams{
		AmountRequested: info.Vested,
	})
	if err != nil {
		return
	}
	msg := &types.Message{
		To:     builtin.VestingActorAddr,
		From:   from,
		Value:  types.NewInt(0),
		Method: builtin.MethodsVesting.WithdrawBalance,
		Params: params,
	}
	c, err := w.sendMessage(node, msg)
	if err != nil {
		return "", err
	}
	return c.String(), nil
}

func (w *Wallet) fullAPI() (fullAPI api.FullNode, closer jsonrpc.ClientCloser, err error) {
	return client.NewFullNodeRPC(context.Background(), w.rpcURL, w.header)
}

//CreateExpert 创建领域专家
func (w *Wallet) CreateExpert(applicationHash string) (expertID string, err error) {
	fmt.Println("Creating expert message")

	owner, err := w.epikWallet.GetDefault(context.Background())
	params, err := actors.SerializeParams(&expertfund.ApplyForExpertParams{
		Owner:           owner,
		ApplicationHash: applicationHash,
	})
	if err != nil {
		return "", err
	}

	msg := &types.Message{
		To:    builtin.ExpertFundActorAddr,
		From:  owner,
		Value: big.Int(types.MustParseEPK("99EPK")),

		Method: builtin.MethodsExpertFunds.ApplyForExpert,
		Params: params,
	}
	fullAPI, closer, err := client.NewFullNodeRPC(context.Background(), w.rpcURL, w.header)
	if err != nil {
		return
	}
	defer closer()
	c, err := w.sendMessage(fullAPI, msg)
	if err != nil {
		return "", err
	}
	lu, err := fullAPI.StateWaitMsg(context.Background(), c, 1)
	if err != nil {
		return
	}
	if lu.Receipt.ExitCode.IsSuccess() {
		var retval expertfund.ApplyForExpertReturn
		err = retval.UnmarshalCBOR(bytes.NewReader(lu.Receipt.Return))
		if err != nil {
			return
		}
		return retval.IDAddress.String(), nil
	}
	return "", fmt.Errorf(lu.Receipt.ExitCode.Error())
}

func (w *Wallet) ExpertNominate(_expert, target string) (cID string, err error) {
	fmt.Println("Nominate expert message")

	owner, err := w.epikWallet.GetDefault(context.Background())
	targetAddr, err := address.NewFromString(target)
	if err != nil {
		return "", err
	}
	expertAddr, err := address.NewFromString(_expert)
	if err != nil {
		return "", err
	}
	params, err := actors.SerializeParams(&targetAddr)
	if err != nil {
		return "", err
	}

	msg := &types.Message{
		To:     expertAddr,
		From:   owner,
		Value:  abi.NewTokenAmount(0),
		Method: expert.Methods.Nominate,
		Params: params,
	}
	fullAPI, closer, err := client.NewFullNodeRPC(context.Background(), w.rpcURL, w.header)
	if err != nil {
		return
	}
	defer closer()
	c, err := w.sendMessage(fullAPI, msg)
	if err != nil {
		return "", err
	}

	return c.String(), nil
}

//ExpertInfo 专家信息
func (w *Wallet) ExpertInfo(addr string) (infoJSON string, err error) {
	expertAddr, err := address.NewFromString(addr)
	if err != nil {
		return
	}
	fullAPI, closer, err := client.NewFullNodeRPC(context.Background(), w.rpcURL, w.header)
	if err != nil {
		return
	}
	defer closer()
	info, err := fullAPI.StateExpertInfo(context.Background(), expertAddr, types.EmptyTSK)
	if err != nil {
		return "", err
	}
	data, err := json.Marshal(info)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

//ExpertList ...
func (w *Wallet) ExpertList() (listJSON string, err error) {
	fullAPI, closer, err := client.NewFullNodeRPC(context.Background(), w.rpcURL, w.header)
	if err != nil {
		return
	}
	defer closer()
	list, err := fullAPI.StateListExperts(context.Background(), types.EmptyTSK)
	if err != nil {
		return
	}
	data, err := json.Marshal(list)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

//VoteSend 投票
func (w *Wallet) VoteSend(candidate string, amount string) (cidStr string, err error) {
	candidateAddr, err := address.NewFromString(candidate)
	if err != nil {
		return
	}
	val, err := types.ParseEPK(amount)
	if err != nil {
		return
	}
	sp, err := actors.SerializeParams(&candidateAddr)
	if err != nil {
		return
	}
	from, err := w.epikWallet.GetDefault(context.Background())
	fullAPI, closer, err := client.NewFullNodeRPC(context.Background(), w.rpcURL, w.header)
	if err != nil {
		return
	}
	defer closer()

	msg := &types.Message{
		From:   from,
		To:     builtin.VoteFundActorAddr,
		Value:  types.BigInt(val),
		Method: builtin.MethodsVote.Vote,
		Params: sp,
	}
	c, err := w.sendMessage(fullAPI, msg)
	if err != nil {
		return "", err
	}
	return c.String(), nil
}

//VoteRescind 撤销
func (w *Wallet) VoteRescind(candidate string, amount string) (cidStr string, err error) {
	candidateAddr, err := address.NewFromString(candidate)
	if err != nil {
		return
	}
	val, err := types.ParseEPK(amount)
	if err != nil {
		return
	}
	p := vote.RescindParams{
		Candidate: candidateAddr,
		Votes:     types.BigInt(val),
	}
	sp, err := actors.SerializeParams(&p)
	if err != nil {
		return
	}
	from, err := w.epikWallet.GetDefault(context.Background())
	fullAPI, closer, err := client.NewFullNodeRPC(context.Background(), w.rpcURL, w.header)
	if err != nil {
		return
	}
	defer closer()

	msg := &types.Message{
		From:   from,
		To:     builtin.VoteFundActorAddr,
		Value:  big.Zero(),
		Method: builtin.MethodsVote.Rescind,
		Params: sp,
	}
	c, err := w.sendMessage(fullAPI, msg)
	if err != nil {
		return "", err
	}
	return c.String(), nil
}

//VoteWithdraw 提现
func (w *Wallet) VoteWithdraw(to string) (cidStr string, err error) {

	var toAddr address.Address
	if to == "" {
		toAddr, err = w.epikWallet.GetDefault(context.Background())
	} else {
		toAddr, err = address.NewFromString(to)
	}
	if err != nil {
		return
	}

	sp, err := actors.SerializeParams(&toAddr)
	if err != nil {
		return "", fmt.Errorf("serializing params: %w", err)
	}
	from, err := w.epikWallet.GetDefault(context.Background())
	fullAPI, closer, err := client.NewFullNodeRPC(context.Background(), w.rpcURL, w.header)
	if err != nil {
		return
	}
	defer closer()

	msg := &types.Message{
		To:     builtin.VoteFundActorAddr,
		From:   from,
		Value:  big.Zero(),
		Method: builtin.MethodsVote.Withdraw,
		Params: sp,
	}
	c, err := w.sendMessage(fullAPI, msg)
	if err != nil {
		return "", err
	}
	return c.String(), nil
}

//VoterInfo 投票信息
func (w *Wallet) VoterInfo(addr string) (infoJSON string, err error) {
	ad, err := address.NewFromString(addr)
	if err != nil {
		return
	}
	fullAPI, closer, err := client.NewFullNodeRPC(context.Background(), w.rpcURL, w.header)
	if err != nil {
		return
	}
	defer closer()
	info, err := fullAPI.StateVoterInfo(context.Background(), ad, types.EmptyTSK)
	if err != nil {
		return
	}
	data, err := json.Marshal(info)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (w *Wallet) MinerInfo(minerID string) (infoJSON string, err error) {
	defer func() {
		if err := recover(); err != nil {
			err = fmt.Errorf("crashed:%+v", err)
		}
	}()
	minerAddr, err := address.NewFromString(minerID)
	if err != nil {
		return
	}
	ctx := context.Background()
	fullAPI, closer, err := client.NewFullNodeRPC(ctx, w.rpcURL, w.header)
	if err != nil {
		return
	}
	defer closer()
	head, err := fullAPI.ChainHead(ctx)
	if err != nil {
		return
	}
	info, err := fullAPI.StateMinerInfo(ctx, minerAddr, types.EmptyTSK)
	if err != nil {
		return
	}
	power, err := fullAPI.StateMinerPower(ctx, minerAddr, types.EmptyTSK)
	if err != nil {
		power = &api.MinerPower{}
	}
	funds, err := fullAPI.StateMinerFunds(ctx, minerAddr, types.EmptyTSK)
	if err != nil {
		return
	}
	fmt.Println(funds)
	retrieve, err := fullAPI.StateRetrievalPledge(ctx, info.Owner, types.EmptyTSK)
	if err != nil {
		retrieve = &api.RetrievalState{
			Balance:   abi.NewTokenAmount(0),
			DayExpend: abi.NewTokenAmount(0),
		}
	}
	coinbase, err := fullAPI.StateCoinbase(ctx, info.Coinbase, types.EmptyTSK)
	if err != nil {
		coinbase = &vesting2.CoinbaseInfo{
			Total:   abi.NewTokenAmount(0),
			Vested:  abi.NewTokenAmount(0),
			Vesting: abi.NewTokenAmount(0),
		}
	}
	myPledge := decimal.Zero
	myRetrieved := decimal.Zero
	retrieveInfo := &api.RetrievalPledgeInfo{}
	mine, err := w.epikWallet.GetDefault(ctx)
	if err == nil {
		myID, err := fullAPI.StateLookupID(ctx, mine, types.EmptyTSK)
		if err == nil {
			pledge, ok := funds.MiningPledgors[myID.String()]
			if ok {
				myPledge = decimal.NewFromBigInt(pledge.Int, -18)
			}

		}
		retrieveInfo, err = fullAPI.StateRetrievalPledgeFrom(ctx, mine, types.EmptyTSK)
		if err == nil {
			am, ok := retrieveInfo.Pledges[info.Owner.String()]
			if ok {
				myRetrieved = decimal.NewFromBigInt(am.Int, -18)
			}
		} else {
			retrieveInfo = &api.RetrievalPledgeInfo{
				Locked:        abi.NewTokenAmount(0),
				UnlockedEpoch: abi.ChainEpoch(0),
			}
		}
	}

	mInfo := &struct {
		CoinBase                string          `json:"coin_base"`
		Owner                   string          `json:"owner"`
		Worker                  string          `json:"worker"`
		MiningPower             decimal.Decimal `json:"mining_power"`
		TotalPower              decimal.Decimal `json:"total_power"`
		CoinbaseBalance         decimal.Decimal `json:"coinbase_balance"`
		Vesting                 decimal.Decimal `json:"vesting"`
		Vested                  decimal.Decimal `json:"vested"`
		MiningPledged           decimal.Decimal `json:"mining_pledged"`
		MyMiningPledge          decimal.Decimal `json:"my_mining_pledge"`
		RetrieveBalance         decimal.Decimal `json:"retrieve_balance"`
		MyRetrievePledge        decimal.Decimal `json:"my_retrieve_pledge"`
		RetrieveLocked          decimal.Decimal `json:"retrieve_locked"`
		RetrieveUnlockEpochLeft decimal.Decimal `json:"retrieve_unlock_epoch_left"`
		RetrieveDayExpend       decimal.Decimal `json:"retrieve_day_expend"`
	}{
		CoinBase:                info.Coinbase.String(),
		Owner:                   info.Owner.String(),
		Worker:                  info.Worker.String(),
		MiningPower:             decimal.NewFromBigInt(power.MinerPower.QualityAdjPower.Int, 0),
		TotalPower:              decimal.NewFromBigInt(power.TotalPower.QualityAdjPower.Int, 0),
		CoinbaseBalance:         decimal.NewFromBigInt(coinbase.Total.Int, -18),
		Vesting:                 decimal.NewFromBigInt(coinbase.Vesting.Int, -18),
		Vested:                  decimal.NewFromBigInt(coinbase.Vested.Int, -18),
		MiningPledged:           decimal.NewFromBigInt(funds.MiningPledge.Int, -18),
		MyMiningPledge:          myPledge,
		RetrieveBalance:         decimal.NewFromBigInt(retrieve.Balance.Int, -18),
		MyRetrievePledge:        myRetrieved,
		RetrieveLocked:          decimal.NewFromBigInt(retrieveInfo.Locked.Int, -18),
		RetrieveUnlockEpochLeft: decimal.NewFromInt(int64(retrieveInfo.UnlockedEpoch - head.Height())),
		RetrieveDayExpend:       decimal.NewFromBigInt(retrieve.DayExpend.Int, -18),
	}

	data, err := json.Marshal(mInfo)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (w *Wallet) MinerPledgeAdd(toMinerID string, amount string) (cidStr string, err error) {
	var toAddr address.Address
	if toMinerID == "" {
		return "", fmt.Errorf("toMinerID is empty")
	}
	toAddr, err = address.NewFromString(toMinerID)
	if err != nil {
		return
	}
	am, err := types.ParseEPK(amount)
	if err != nil {
		return "", err
	}
	from, err := w.epikWallet.GetDefault(context.Background())
	if err != nil {
		return "", err
	}
	fullAPI, closer, err := client.NewFullNodeRPC(context.Background(), w.rpcURL, w.header)
	if err != nil {
		return
	}
	defer closer()
	bal, err := fullAPI.WalletBalance(context.Background(), from)
	if err != nil {
		return
	}
	if bal.LessThan(big.Int(am)) {
		return "", fmt.Errorf("not enough balance")
	}
	msg := &types.Message{
		To:     toAddr,
		From:   from,
		Value:  abi.TokenAmount(am),
		Method: miner.Methods.AddPledge,
		Params: nil,
	}
	c, err := w.sendMessage(fullAPI, msg)
	if err != nil {
		return "", err
	}
	return c.String(), nil
}

func (w *Wallet) MinerPledgeOneClick(minerStr string) (err error) {
	minerStr = strings.TrimSpace(minerStr)
	if minerStr == "" {
		return fmt.Errorf("miners is empty")
	}
	minerIDs := strings.Split(minerStr, ",")
	am, err := types.ParseEPK("1000EPK")
	if err != nil {
		return err
	}
	from, err := w.epikWallet.GetDefault(context.Background())
	if err != nil {
		return err
	}
	node, closer, err := client.NewFullNodeRPC(context.Background(), w.rpcURL, w.header)
	if err != nil {
		return
	}
	defer closer()
	bal, err := node.WalletBalance(context.Background(), from)
	if err != nil {
		return
	}
	if bal.LessThan(big.Mul(big.Int(am), big.NewInt(int64(len(minerIDs))))) {
		return fmt.Errorf("not enough balance")
	}
	for _, minerID := range minerIDs {
		m, err := address.NewFromString(minerID)
		if err != nil {
			return err
		}
		msg := &types.Message{
			To:     m,
			From:   from,
			Value:  abi.TokenAmount(am),
			Method: miner.Methods.AddPledge,
			Params: nil,
		}
		_, err = w.sendMessage(node, msg)
		if err != nil {
			return err
		}
	}
	return
}

func (w *Wallet) MinerPledgeApplyWithdrawOneClick(minerStr string) (err error) {
	minerStr = strings.TrimSpace(minerStr)
	if minerStr == "" {
		return fmt.Errorf("miners is empty")
	}
	minerIDs := strings.Split(minerStr, ",")

	from, err := w.epikWallet.GetDefault(context.Background())
	if err != nil {
		return err
	}
	node, closer, err := client.NewFullNodeRPC(context.Background(), w.rpcURL, w.header)
	if err != nil {
		return
	}
	defer closer()
	fromID, err := node.StateLookupID(context.Background(), from, types.EmptyTSK)
	if err != nil {
		return
	}
	for _, minerID := range minerIDs {
		m, err := address.NewFromString(minerID)
		if err != nil {
			continue
		}
		funds, err := node.StateMinerFunds(context.Background(), m, types.EmptyTSK)
		if err != nil {
			continue
		}
		pledged, ok := funds.MiningPledgors[fromID.String()]
		if !ok {
			continue
		}
		params, err := actors.SerializeParams(&fminer.WithdrawPledgeParams{
			AmountRequested: abi.TokenAmount(pledged), // Default to attempting to withdraw all the extra funds in the miner actor
		})
		if err != nil {
			continue
		}
		msg := &types.Message{
			To:     m,
			From:   from,
			Value:  types.NewInt(0),
			Method: miner.Methods.ApplyForWithdraw,
			Params: params,
		}
		_, err = w.sendMessage(node, msg)
		if err != nil {
			continue
		}
	}
	return
}

func (w *Wallet) MinerPledgeWithdrawOneClick(minerStr string) (err error) {
	minerStr = strings.TrimSpace(minerStr)
	if minerStr == "" {
		return fmt.Errorf("miners is empty")
	}
	minerIDs := strings.Split(minerStr, ",")

	from, err := w.epikWallet.GetDefault(context.Background())
	if err != nil {
		return err
	}
	node, closer, err := client.NewFullNodeRPC(context.Background(), w.rpcURL, w.header)
	if err != nil {
		return
	}
	defer closer()
	ts, err := node.ChainHead(context.Background())
	if err != nil {
		return
	}
	fromID, err := node.StateLookupID(context.Background(), from, types.EmptyTSK)
	if err != nil {
		return
	}
	for _, minerID := range minerIDs {
		m, err := address.NewFromString(minerID)
		if err != nil {
			continue
		}
		funds, err := node.StateMinerFunds(context.Background(), m, types.EmptyTSK)
		if err != nil {
			continue
		}
		locked, ok := funds.MiningPledgeLocked[fromID.String()]
		if !ok {
			err = errors.New("no pledged balance")
			continue
		}
		if locked.EffectiveAt > ts.Height() {
			fmt.Println(locked.EffectiveAt)
			err = errors.New("no unlocked time")
			continue
		}
		params, err := actors.SerializeParams(&fminer.WithdrawPledgeParams{
			AmountRequested: abi.TokenAmount(locked.Amount),
		})

		if err != nil {
			continue
		}
		msg := &types.Message{
			To:     m,
			From:   from,
			Value:  types.NewInt(0),
			Method: miner.Methods.WithdrawPledge,
			Params: params,
		}
		_, err = w.sendMessage(node, msg)
		if err != nil {
			continue
		}
	}
	return
}

func (w *Wallet) MinerPledgeWithdraw(toMinerID string, amount string) (cidStr string, err error) {
	var toAddr address.Address
	if toMinerID == "" {
		return "", fmt.Errorf("toMinerID is empty")
	}
	toAddr, err = address.NewFromString(toMinerID)
	if err != nil {
		return
	}
	am, err := types.ParseEPK(amount)
	if err != nil {
		return "", err
	}
	from, err := w.epikWallet.GetDefault(context.Background())
	node, closer, err := client.NewFullNodeRPC(context.Background(), w.rpcURL, w.header)
	if err != nil {
		return
	}
	defer closer()
	fromID, err := node.StateLookupID(context.Background(), from, types.EmptyTSK)
	if err != nil {
		return
	}
	funds, err := node.StateMinerFunds(context.Background(), toAddr, types.EmptyTSK)
	if err != nil {
		return
	}
	locked, ok := funds.MiningPledgeLocked[fromID.String()]
	if !ok {
		err = errors.New("no pledged balance")
		return
	}
	if locked.Amount.LessThan(big.Int(am)) {
		return "", fmt.Errorf("not enough balance")
	}
	ts, err := node.ChainHead(context.Background())
	if err != nil {
		return
	}
	if locked.EffectiveAt > ts.Height() {
		fmt.Println(locked.EffectiveAt)
		err = errors.New("no unlocked time")
		return
	}
	params, err := actors.SerializeParams(&fminer.WithdrawPledgeParams{
		AmountRequested: abi.TokenAmount(am), // Default to attempting to withdraw all the extra funds in the miner actor
	})
	if err != nil {
		return "", err
	}

	msg := &types.Message{
		To:     toAddr,
		From:   from,
		Value:  types.NewInt(0),
		Method: miner.Methods.WithdrawPledge,
		Params: params,
	}
	c, err := w.sendMessage(node, msg)
	if err != nil {
		return "", err
	}
	return c.String(), nil
}

func (w *Wallet) MinerPledgeApplyWithdraw(minerID string) (cidStr string, err error) {
	minerAddr, err := address.NewFromString(minerID)
	if err != nil {
		return
	}
	node, closer, err := client.NewFullNodeRPC(context.Background(), w.rpcURL, w.header)
	if err != nil {
		return
	}
	defer closer()
	from, err := w.epikWallet.GetDefault(context.Background())
	if err != nil {
		return
	}
	fromID, err := node.StateLookupID(context.Background(), from, types.EmptyTSK)
	if err != nil {
		return
	}
	funds, err := node.StateMinerFunds(context.Background(), minerAddr, types.EmptyTSK)
	if err != nil {
		return
	}
	pledged, ok := funds.MiningPledgors[fromID.String()]
	if !ok {
		return "", fmt.Errorf("no pledged balance")
	}
	params, err := actors.SerializeParams(&fminer.WithdrawPledgeParams{
		AmountRequested: abi.TokenAmount(pledged), // Default to attempting to withdraw all the extra funds in the miner actor
	})
	if err != nil {
		return
	}
	msg := &types.Message{
		To:     minerAddr,
		From:   from,
		Value:  types.NewInt(0),
		Method: miner.Methods.ApplyForWithdraw,
		Params: params,
	}
	c, err := w.sendMessage(node, msg)
	if err != nil {
		return "", err
	}
	return c.String(), nil
}

func (w *Wallet) MinerPledgeTransfer(fromMinerID, toMinerID string, amount string) (cidStr string, err error) {
	fromMiner, err := address.NewFromString(fromMinerID)
	if err != nil {
		return
	}
	toMiner, err := address.NewFromString(toMinerID)
	if err != nil {
		return
	}
	am, err := types.ParseEPK(amount)
	if err != nil {
		return "", err
	}
	node, closer, err := client.NewFullNodeRPC(context.Background(), w.rpcURL, w.header)
	if err != nil {
		return
	}
	defer closer()
	from, err := w.epikWallet.GetDefault(context.Background())
	if err != nil {
		return
	}

	params, err := actors.SerializeParams(&fminer.TransferPledgeParamsV2{
		Miner:  toMiner,
		Amount: abi.TokenAmount(am),
	})
	if err != nil {
		return
	}
	msg := &types.Message{
		To:     fromMiner,
		From:   from,
		Value:  types.NewInt(0),
		Method: miner.Methods.TransferPledgeV2,
		Params: params,
	}
	c, err := w.sendMessage(node, msg)
	if err != nil {
		return "", err
	}
	return c.String(), nil
}

func (w *Wallet) RetrievePledgeState(addr string) (stateJSON string, err error) {
	target, err := address.NewFromString(addr)
	if err != nil {
		return
	}

	node, closer, err := client.NewFullNodeRPC(context.Background(), w.rpcURL, w.header)
	if err != nil {
		return
	}
	defer closer()
	if target.Protocol() == address.ID {
		target, err = node.StateAccountKey(context.Background(), target, types.EmptyTSK)
		if err != nil {
			return
		}
	}
	state, err := node.StateRetrievalPledgeFrom(context.Background(), target, types.EmptyTSK)
	if err != nil {
		return
	}
	data, err := json.Marshal(state)
	if err != nil {
		return
	}
	return string(data), nil
}

func (w *Wallet) RetrievePledgeAdd(target string, miner string, amount string) (cidStr string, err error) {
	var targetAddr address.Address
	var minerAddr address.Address
	if target == "" {
		return "", fmt.Errorf("target is empty")
	}
	targetAddr, err = address.NewFromString(target)
	if err != nil {
		return
	}

	am, err := types.ParseEPK(amount)
	if err != nil {
		return "", err
	}
	fullAPI, closer, err := client.NewFullNodeRPC(context.Background(), w.rpcURL, w.header)
	if err != nil {
		return
	}
	defer closer()
	if targetAddr.Protocol() == address.ID {
		targetAddr, err = fullAPI.StateAccountKey(context.Background(), targetAddr, types.EmptyTSK)
		if err != nil {
			return
		}
	}
	minerAddr, err = address.NewFromString(miner)
	var params []byte
	if err == nil && !minerAddr.Empty() {
		params, err = actors.SerializeParams(&retrieval.PledgeParams{
			Address: targetAddr,
			Miners:  []address.Address{minerAddr},
		})
	} else {
		params, err = actors.SerializeParams(&retrieval.PledgeParams{
			Address: targetAddr,
			Miners:  []address.Address{},
		})
	}
	if err != nil {
		return "", xerrors.Errorf("serializing params failed: %w", err)
	}
	from, err := w.epikWallet.GetDefault(context.Background())

	bal, err := fullAPI.WalletBalance(context.Background(), from)
	if err != nil {
		return
	}
	if bal.LessThan(big.Int(am)) {
		return "", fmt.Errorf("not enough balance")
	}
	msg := &types.Message{
		To:     retrieval.Address,
		From:   from,
		Value:  big.Int(am),
		Method: retrieval.Methods.Pledge,
		Params: params,
	}
	c, err := w.sendMessage(fullAPI, msg)
	if err != nil {
		return "", err
	}
	return c.String(), nil
}

func (w *Wallet) RetrievePledgeBind(miner string, amount string) (cidStr string, err error) {
	var minerAddr address.Address

	if miner == "" {
		return "", fmt.Errorf("target is empty")
	}
	minerAddr, err = address.NewFromString(miner)
	if err != nil {
		return
	}
	am, err := types.ParseEPK(amount)
	if err != nil {
		return "", err
	}
	from, err := w.epikWallet.GetDefault(context.Background())
	if err != nil {
		return "", err
	}
	params, err := actors.SerializeParams(&retrieval.BindMinersParams{
		Pledger: from,
		Miners:  []address.Address{minerAddr},
	})
	if err != nil {
		return "", xerrors.Errorf("serializing params failed: %w", err)
	}

	fullAPI, closer, err := client.NewFullNodeRPC(context.Background(), w.rpcURL, w.header)
	if err != nil {
		return
	}
	defer closer()
	bal, err := fullAPI.WalletBalance(context.Background(), from)
	if err != nil {
		return
	}
	if bal.LessThan(big.Int(am)) {
		return "", fmt.Errorf("not enough balance")
	}
	msg := &types.Message{
		To:     retrieval.Address,
		From:   from,
		Value:  big.Zero(),
		Method: retrieval.Methods.BindMiners,
		Params: params,
	}
	c, err := w.sendMessage(fullAPI, msg)
	if err != nil {
		return "", err
	}
	return c.String(), nil
}

func (w *Wallet) RetrievePledgeUnBind(miner string, amount string) (cidStr string, err error) {
	var minerAddr address.Address

	if miner == "" {
		return "", fmt.Errorf("target is empty")
	}
	minerAddr, err = address.NewFromString(miner)
	if err != nil {
		return
	}
	am, err := types.ParseEPK(amount)
	if err != nil {
		return "", err
	}
	from, err := w.epikWallet.GetDefault(context.Background())
	if err != nil {
		return "", err
	}
	params, err := actors.SerializeParams(&retrieval.BindMinersParams{
		Pledger: from,
		Miners:  []address.Address{minerAddr},
	})
	if err != nil {
		return "", xerrors.Errorf("serializing params failed: %w", err)
	}

	fullAPI, closer, err := client.NewFullNodeRPC(context.Background(), w.rpcURL, w.header)
	if err != nil {
		return
	}
	defer closer()
	bal, err := fullAPI.WalletBalance(context.Background(), from)
	if err != nil {
		return
	}
	if bal.LessThan(big.Int(am)) {
		return "", fmt.Errorf("not enough balance")
	}
	msg := &types.Message{
		To:     retrieval.Address,
		From:   from,
		Value:  big.Zero(),
		Method: retrieval.Methods.UnbindMiners,
		Params: params,
	}
	c, err := w.sendMessage(fullAPI, msg)
	if err != nil {
		return "", err
	}
	return c.String(), nil
}

func (w *Wallet) RetrievePledgeApplyWithdraw(target string, amount string) (cidStr string, err error) {
	if target == "" {
		return "", fmt.Errorf("toMinerID is empty")
	}
	targetAddr, err := address.NewFromString(target)
	if err != nil {
		return
	}
	am, err := types.ParseEPK(amount)
	if err != nil {
		return "", err
	}

	fullAPI, closer, err := client.NewFullNodeRPC(context.Background(), w.rpcURL, w.header)
	if err != nil {
		return
	}
	defer closer()
	if targetAddr.Protocol() == address.ID {
		targetAddr, err = fullAPI.StateAccountKey(context.Background(), targetAddr, types.EmptyTSK)
		if err != nil {
			return "", err
		}
	}
	params, err := actors.SerializeParams(&retrieval.WithdrawBalanceParams{
		Target: targetAddr,
		Amount: big.Int(am),
	})
	if err != nil {
		return "", xerrors.Errorf("serializing params failed: %w", err)
	}
	from, err := w.epikWallet.GetDefault(context.Background())
	if err != nil {
		return
	}
	msg := &types.Message{
		To:     retrieval.Address,
		From:   from,
		Value:  abi.NewTokenAmount(0),
		Method: retrieval.Methods.ApplyForWithdraw,
		Params: params,
	}
	c, err := w.sendMessage(fullAPI, msg)
	if err != nil {
		return "", err
	}
	return c.String(), nil
}
func (w *Wallet) RetrievePledgeWithdraw(amount string) (cidStr string, err error) {

	am, err := types.ParseEPK(amount)
	if err != nil {
		return "", err
	}
	from, err := w.epikWallet.GetDefault(context.Background())
	if err != nil {
		return "", err
	}
	amtoken := abi.TokenAmount(am)
	params, err := actors.SerializeParams(&amtoken)
	if err != nil {
		return "", xerrors.Errorf("serializing params failed: %w", err)
	}

	fullAPI, closer, err := client.NewFullNodeRPC(context.Background(), w.rpcURL, w.header)
	if err != nil {
		return
	}
	defer closer()

	msg := &types.Message{
		To:     retrieval.Address,
		From:   from,
		Value:  abi.NewTokenAmount(0),
		Method: retrieval.Methods.WithdrawBalance,
		Params: params,
	}
	c, err := w.sendMessage(fullAPI, msg)
	if err != nil {
		return "", err
	}
	return c.String(), nil
}
