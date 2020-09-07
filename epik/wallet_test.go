package epik

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/EpiK-Protocol/epik-wallet-golib/hd"
	"github.com/hashicorp/go-uuid"
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
	epikAddr, err := epikWallet.GenerateKey("secp256k1", seed, "m/44'/3924011'/1'/0/0")
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

func TestSendEPK(t *testing.T) {
	epikWallet, err := NewWallet()
	panicErr(err)
	addr, err := epikWallet.Import("7b2254797065223a22626c73222c22507269766174654b6579223a2231386372674765746b4257634d7155654659304545426d3261375632412f53467a30315079425434596c673d227d")
	panicErr(err)
	fmt.Println(addr)
	err = epikWallet.SetRPC("ws://171.221.243.41:1234/rpc/v0", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJBbGxvdyI6WyJyZWFkIiwid3JpdGUiLCJzaWduIiwiYWRtaW4iXX0.7bKI-eB5MByU3BjLH7_Al4ZYgLseKibKWSwURDu_t9k")
	panicErr(err)
	balance, err := epikWallet.Balance(addr)
	panicErr(err)
	fmt.Println(balance)
	t.Log(balance)
	cid, err := epikWallet.Send("t1u3atfaybl4r3lnun3waojvyrsnulgfqtdn3vbry", "1")
	panicErr(err)
	t.Log(cid)
	msgs, err := epikWallet.MessageList(0, addr)
	panicErr(err)
	t.Log(msgs)
}

func TestMessages(t *testing.T) {
	epikWallet, err := NewWallet()
	panicErr(err)
	addr, err := epikWallet.Import("7b2254797065223a22626c73222c22507269766174654b6579223a2231386372674765746b4257634d7155654659304545426d3261375632412f53467a30315079425434596c673d227d")
	panicErr(err)
	fmt.Println(addr)
	err = epikWallet.SetRPC("ws://18.181.234.52:1234/rpc/v0", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJBbGxvdyI6WyJyZWFkIiwid3JpdGUiXX0.6lL7ayYWfLqEh0BqOtCwUvLVEJ5LJ1BMb3HFRRaHsVY")
	panicErr(err)
	balance, err := epikWallet.Balance("t3sbsmbkmvu7kpc5adrcaubyuna33t3gl2medbcjo65ttcuppb45srztz52lgy4iqrf3nwle2okvfhlf5xddgq")
	panicErr(err)
	fmt.Println(balance)
	msgs, err := epikWallet.MessageList(0, "t3sbsmbkmvu7kpc5adrcaubyuna33t3gl2medbcjo65ttcuppb45srztz52lgy4iqrf3nwle2okvfhlf5xddgq")
	panicErr(err)
	t.Log(msgs)
}

func panicErr(err error) {
	if err != nil {
		panic(err)
	}
}

func TestPathSeed(t *testing.T) {
	// mne, err := hdwallet.NewMnemonic(128)
	// panicErr(err)
	// t.Log(mne)
	seed, err := hdwallet.NewSeedFromMnemonic("fine bubble drum remember motor kiss arctic leisure adjust immune involve expect")
	panicErr(err)
	t.Log(seed)
	wallet, err := NewWallet()
	panicErr(err)
	addr, err := wallet.GenerateKey("bls", seed, "m/44'/3924011'/1'/0/0")
	panicErr(err)
	fmt.Println(addr)
	err = wallet.SetRPC("ws://18.181.234.52:1234/rpc/v0", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJBbGxvdyI6WyJyZWFkIiwid3JpdGUiLCJzaWduIiwiYWRtaW4iXX0.67ImUqjhRCP7gNlpFgHc4M4wjNC30gUfQDSTuQQu428")
	panicErr(err)
	bal, err := wallet.Balance("t3sudvhdhpem6pgl4hksbj5c4adqvwaxh6lku5iwtwquvortz33tndp5zfgmrfzvv2h4g6p724qfh5svunfoga")
	panicErr(err)
	fmt.Println(bal)
}

func TestExport(t *testing.T) {
	seed, err := hdwallet.NewSeedFromMnemonic("theme bacon syrup naive flock puppy found boss hurdle crisp grit oak")
	panicErr(err)
	t.Log(seed)
	hdWallet, err := hd.NewFromSeed(seed)
	addr, err := hdWallet.Derive("m/44'/60'/0'/0/0", true)
	t.Log(addr)
	wallet, err := NewWallet()
	panicErr(err)
	addr, err = wallet.GenerateKey("bls", seed, "m/44'/3924011'/1'/0/0")
	panicErr(err)
	t.Log(addr)
	privateKey, err := wallet.Export(addr)
	panicErr(err)
	t.Log(privateKey)

}

func TestImport(t *testing.T) {

}

func TestBalance(t *testing.T) {
	addr := "t3qkoospzrp6videditkj5imscrf33shdncxy7ukd4klwbpmcg4gm3olxuxykixolhg3w2vurwp2u4qibx7wvq"
	wallet, err := NewWallet()
	panicErr(err)
	err = wallet.SetRPC("ws://18.181.234.52:1234/rpc/v0", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJBbGxvdyI6WyJyZWFkIiwid3JpdGUiXX0.6lL7ayYWfLqEh0BqOtCwUvLVEJ5LJ1BMb3HFRRaHsVY")
	panicErr(err)
	balance, err := wallet.Balance(addr)
	panicErr(err)
	t.Log(balance)
	msgs, err := wallet.MessageList(0, addr)
	panicErr(err)
	t.Log(msgs)
}

func TestGenRobotAccounts(t *testing.T) {
	type Miner struct {
		ID             string    `json:"id" gorm:"primarykey"`
		WeiXin         string    `json:"wei_xin" gorm:"unique"`
		Mnemonic       string    `json:"mnemonic"`
		EpikAddress    string    `json:"epik_address"`
		Erc20Address   string    `json:"erc20_address"`
		EpikPrivateKey string    `json:"epik_privatekey"`
		Status         string    `json:"status"`
		CreatedAt      time.Time `json:"created_at"`
	}
	miners := []*Miner{}
	for i := 0; i < 1000; i++ {
		miner := &Miner{}
		mnemonic, err := hd.NewMnemonic(128)
		panicErr(err)
		// t.Log(mnemonic)
		miner.Mnemonic = mnemonic
		seed, err := hd.SeedFromMnemonic(mnemonic)
		panicErr(err)
		// t.Log(seed)
		hdWallet, err := hd.NewFromSeed(seed)
		panicErr(err)
		hdAddr, err := hdWallet.Derive("m/44'/60'/0'/0/0", true)
		panicErr(err)
		miner.Erc20Address = hdAddr
		epikWallet, err := NewWallet()
		panicErr(err)
		epkAddr, err := epikWallet.GenerateKey("bls", seed, "m/44'/3924011'/1'/0/0")
		panicErr(err)
		miner.EpikAddress = epkAddr
		epikPK, err := epikWallet.Export(epkAddr)
		panicErr(err)
		miner.EpikPrivateKey = epikPK
		miner.Status = "confirmed"
		miner.CreatedAt = time.Now()
		miner.WeiXin = fmt.Sprintf("kt%04d", i)
		miner.ID, err = uuid.GenerateUUID()
		panicErr(err)
		miners = append(miners, miner)
		// t.Log(miner)
	}

	jsonFile, err := os.OpenFile("./kt.json", os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0777)
	panicErr(err)
	data, err := json.Marshal(miners)
	panicErr(err)
	_, err = jsonFile.Write(data)
	panicErr(err)
	jsonFile.Close()

	sqlFile, err := os.OpenFile("./kt.sql", os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0777)
	panicErr(err)
	_, err = sqlFile.WriteString("INSERT INTO miner (id,wei_xin,epik_address,erc20_address,status,created_at) VALUES \n")
	panicErr(err)
	for _, miner := range miners {
		_, err = sqlFile.WriteString(fmt.Sprintf("('%s','%s','%s','%s','%s','%s'),\n", miner.ID, miner.WeiXin, miner.EpikAddress, miner.Erc20Address, miner.Status, miner.CreatedAt.Format(time.RFC3339)))
		panicErr(err)
	}
	sqlFile.Close()
}
