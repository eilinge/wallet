package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"math/big"

	"crypto/rand"

	"context"

	"wallet/hdwallet"

	"github.com/davecgh/go-spew/spew"

	// "wallet/utils"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/tyler-smith/go-bip39"
)

// LightScryptP  ...
const (
	LightScryptP = 6
	LightScryptN = 1 << 12
)

var (
	auth = "tester"
)

// EncryptedKeyJSONV3 ...
type EncryptedKeyJSONV3 struct {
	Address string              `json:"address"`
	Crypto  keystore.CryptoJSON `json:"crypto"`
	ID      string              `json:"id"`
	Version int                 `json:"version"`
}

func testMnemonic() (cj keystore.CryptoJSON, err error) {

	// [128, 256] 32的整数倍
	b, err := bip39.NewEntropy(128)
	if err != nil {
		log.Panic("failed to NewEntropy:", err)
	}
	// [79 203 129 59 164 87 94 188 137 214 205 242 62 2 27 134]
	fmt.Println(b)

	mnemonic, err := bip39.NewMnemonic(b)
	if err != nil {
		log.Panic("failed to NewMnemonic:", err)
	}

	fmt.Println("mnemonic:", mnemonic)
	// pss: "123"
	seed := bip39.NewSeed(mnemonic, auth)

	// fa8ca27fa578c68f11d1b138c7cc462302fd95ac10d31d2fb92d73bb8de1f3edfd258bf556fe59d1ef55c83f634c03b74b2c5026f770f6b504cf04a9452e1198
	fmt.Printf("%x\n", seed)

	// masterKey, _ := bip32.NewMasterKey(seed)
	// publicKey := masterKey.PublicKey()
	// fmt.Println("Master private key: ", masterKey)
	// fmt.Println("Master public key: ", publicKey)
	cj, err = GetEncryptDataV3(seed, []byte(auth), LightScryptN, LightScryptP)
	return
	// masterKey.MarshalJSON()

}

func test2() {
	mne := "exist foster exclude emerge invest furnace chef super vendor useless manage arrange"

	wallet, err := hdwallet.NewFromMnemonic(mne)
	if err != nil {
		log.Panic(err)
	}

	path := hdwallet.MustParseDerivationPath("m/44'/60'/0'/0/0")
	account, err := wallet.Derive(path, false)
	if err != nil {
		log.Panic(err)
	}
	// 0xb2E4BEec903EDB94054b4f91C1722A691F82a6C6
	fmt.Println(account.Address.Hex())

	path = hdwallet.MustParseDerivationPath("m/44'/60'/0'/0/1")
	account, err = wallet.Derive(path, false)
	if err != nil {
		log.Panic(err)
	}
	// 0x6C226da95295B3B6b32953eC8518C27946bc00E1
	fmt.Println(account.Address.Hex())
}

// TestSign test sign valiable
func TestSign() {
	mnemonic := "exist foster exclude emerge invest furnace chef super vendor useless manage arrange"
	wallet, err := hdwallet.NewFromMnemonic(mnemonic)

	if err != nil {
		log.Fatal(err)
	}

	path := hdwallet.MustParseDerivationPath("m/44'/60'/0'/0/0")

	account, err := wallet.Derive(path, true)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("path:", account.URL.Path)

	nonce := uint64(0)
	value := big.NewInt(1000000000000000000) // 1个以太
	toAddress := common.HexToAddress("0x29155963f8632EaeD108f6A81eA65c75C62e77c0")
	gasLimit := uint64(21000)
	gasPrice := big.NewInt(21000000000)
	var data []byte

	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, data)
	signedTx, err := wallet.SignTx(account, tx, nil)
	if err != nil {
		log.Fatal(err)
	}
	//为了展示
	spew.Dump(signedTx)
}

func testSendTransaction() {
	cli, err := ethclient.Dial("HTTP://127.0.0.1:7545") //注意地址变化 8545
	if err != nil {
		log.Panic(err)
	}

	defer cli.Close()

	mnemonic := "exist foster exclude emerge invest furnace chef super vendor useless manage arrange"
	wallet, err := hdwallet.NewFromMnemonic(mnemonic)
	if err != nil {
		log.Fatal(err)
	}

	path := hdwallet.MustParseDerivationPath("m/44'/60'/0'/0/1") //第2个账户地址
	account, err := wallet.Derive(path, true)
	if err != nil {
		log.Fatal(err)
	}

	// 3000441000000000000
	nonce := uint64(0)
	value := big.NewInt(1000000000000000000)
	toAddress := common.HexToAddress("0x44f4CD617655104649C1b866D20D5EAE198deD38")
	gasLimit := uint64(21000)
	gasPrice := big.NewInt(21000000000)
	var data []byte

	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, data)

	signedTx, err := wallet.SignTx(account, tx, nil)
	if err != nil {
		log.Fatal("failed to signed tx:", err)
	}

	err = cli.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Panic("failed to send transaction:", err)
	}

	// bl, err := cli.BalanceAt(cli, toAddress, big.NewInt(0))
	// fmt.Println(bl.Int64(), err)
}

// GetEncryptDataV3 ...
func GetEncryptDataV3(data, auth []byte, scryptN, scryptP int) (cj keystore.CryptoJSON, err error) {
	cj, err = keystore.EncryptDataV3(data, auth, scryptN, scryptP)

	if err != nil {
		fmt.Println("failed to keystore.EncryptDataV3:", err)
		return
	}
	return
}

// NewKey ...
func NewKey() {
	// UTC--2019-07-04T12-36-09.633463400Z--48b650a7225208e0fb066e0beb04e97391647a0e
	newKey := keystore.NewKeyForDirectICAP(rand.Reader)
	// GetEncryptKey(newKey)
	// key, _ := newKey.MarshalJSON()
	// newKey.UnmarshalJSON(key)

	fmt.Println("newKey: ", newKey.Id)
	cj, _ := testMnemonic()
	// 	type CryptoJSON struct {
	// 	Cipher       string                 `json:"cipher"`
	// 	CipherText   string                 `json:"ciphertext"`
	// 	CipherParams cipherparamsJSON       `json:"cipherparams"`
	// 	KDF          string                 `json:"kdf"`
	// 	KDFParams    map[string]interface{} `json:"kdfparams"`
	// 	MAC          string                 `json:"mac"`
	// }
	fmt.Println("CryptoJSON: ", cj.Cipher)
	// StoreKey(dir, auth string, scryptN, scryptP int)
	addr, _ := keystore.StoreKey(KeyDir, auth, LightScryptN, LightScryptP)
	fmt.Println("addr: ", addr.Hex())
	// NewKeyStore(keydir string, scryptN, scryptP int)
	newStore := keystore.NewKeyStore(KeyDir, LightScryptN, LightScryptP)
	// accounts.Account{Address:[27 119 31 133 241 218 70 70 220 149 222 61 173 42 31 125 4 204 139 217],
	// URL:accounts.URL{Scheme:"keystore", Path:"c:\\c\\Users\\wuchan4x\\Desktop\\eilinge\\goEcho\\src\\weWallet\\UTC
	// --2019-07-12T01-15-32.578987200Z--1b771f85f1da4646dc95de3dad2a1f7d04cc8bd9"}}
	// fmt.Printf("Accounts: %#v \n", newStore.Accounts()[0].URL.Path)
	// 0x1B771F85F1dA4646dc95De3daD2A1f7d04CC8BD9
	// fmt.Printf("Account.address: %#v \n", newStore.Accounts()[0].Address.Hex())
	// fmt.Println("Wallets:", newStore.Wallets())

	filePath := KeyFileName(newStore.Accounts()[0].Address)
	fileName := KeyDir + "/" + filePath
	fmt.Println("fileName: ", fileName)
	// fileName := newStore.Accounts()[0].Address.Hex()

	ej := &EncryptedKeyJSONV3{
		Address: newStore.Accounts()[0].Address.Hex(),
		Crypto:  cj,
		ID:      newKey.Id.String(),
		Version: 3,
	}
	// fmt.Println(fileName, ej)
	// var ejtest []byte
	byEj, _ := json.Marshal(ej)
	// log.Println(ejtest)
	// log.Println(byEj)
	// WriteTemporaryKeyFile(fileName, byEj)
	WriteKeyFile(fileName, byEj)

}

// GetEncryptKey ...
func GetEncryptKey(key *keystore.Key) {
	// (key *Key, auth string, scryptN, scryptP int)
	newkey, _ := keystore.EncryptKey(key, auth, LightScryptN, LightScryptP)
	fmt.Println("newkey: ", string(newkey))
}

func main() {
	// testMnemonic()
	// test2()
	// testSign()
	// testSendTransaction()
	NewKey()
}
