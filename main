package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"crypto/rand"

	"wallet/hdwallet"
	"wallet/utils"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
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

// testMnemonic ...
func testMnemonic() {
	// func testMnemonic() (cj keystore.CryptoJSON, err error, mnemonic string) {

	// [128, 256] 32的整数倍
	b, err := bip39.NewEntropy(128)
	if err != nil {
		log.Panic("failed to NewEntropy:", err)
	}
	// [79 203 129 59 164 87 94 188 137 214 205 242 62 2 27 134]
	// fmt.Println(b)

	mnemonic, err := bip39.NewMnemonic(b)
	if err != nil {
		log.Panic("failed to NewMnemonic:", err)
	}

	fmt.Println("mnemonic:", mnemonic)
	seed := bip39.NewSeed(mnemonic, auth)

	// fa8ca27fa578c68f11d1b138c7cc462302fd95ac10d31d2fb92d73bb8de1f3edfd258bf556fe59d1ef55c83f634c03b74b2c5026f770f6b504cf04a9452e1198
	HdKeyStore(mnemonic, seed)
	// fmt.Printf("%x\n", seed)

	// return

}

// HdKeyStore ...
func HdKeyStore(mne string, seed []byte) {
	wallet, err := hdwallet.NewFromMnemonic(mne)
	if err != nil {
		log.Panic(err)
	}

	for i := 0; i < 1; i++ {
		path := hdwallet.MustParseDerivationPath("m/44'/60'/0'/0/" + strconv.Itoa(i))
		account, err := wallet.Derive(path, false)
		if err != nil {

			log.Println("failed to HDwallet accounts err: ", err)
		}
		// 0xb2E4BEec903EDB94054b4f91C1722A691F82a6C6
		fmt.Println(account.Address.Hex())
		cj, err := GetEncryptDataV3(seed, []byte(auth), LightScryptN, LightScryptP)
		NewKey(account.Address, cj)
	}

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
func NewKey(mnemonic common.Address, cj keystore.CryptoJSON) {
	// UTC--2019-07-04T12-36-09.633463400Z--48b650a7225208e0fb066e0beb04e97391647a0e
	newKey := keystore.NewKeyForDirectICAP(rand.Reader)
	// GetEncryptKey(newKey)
	// key, _ := newKey.MarshalJSON()
	// newKey.UnmarshalJSON(key)

	fmt.Println("newKey: ", newKey.Id)

	// 生成账户对应钱包的数据
	// addr, _ := keystore.StoreKey(utils.KeyDir, auth, LightScryptN, LightScryptP)
	// fmt.Println("addr: ", addr.Hex())

	accountsName := utils.KeyFileName(mnemonic)

	// 替换掉addrees, Crypto, 在写入文件中即可
	ej := &EncryptedKeyJSONV3{
		Address: mnemonic.Hex(),
		Crypto:  cj,
		ID:      newKey.Id.String(),
		Version: 3,
	}

	byEj, _ := json.Marshal(ej)

	err := utils.WriteKeyFile(accountsName, byEj)
	if err != nil {
		log.Fatal("writeKeyFile err")
	}
	log.Println("write key file success...")

}

// GetEncryptKey ...
func GetEncryptKey(key *keystore.Key) {
	// (key *Key, auth string, scryptN, scryptP int)
	newkey, _ := keystore.EncryptKey(key, auth, LightScryptN, LightScryptP)
	fmt.Println("newkey: ", string(newkey))
}

// DecryptoJSON ...
func DecryptoJSON(file string) {
	// keyjson, err := ioutil.ReadFile(file)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// // fmt.Println("keyJson: ", string(keyjson))
	// var ekj EncryptedKeyJSONV3
	// json.Unmarshal(keyjson, &ekj)
	// fmt.Println("account address: ", ekj.Address)

	accountAddr, err := hdwallet.ParseDerivationPath(file)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(accountAddr)
}
func main() {
	// testMnemonic()
	// test2()
	utils.TestSign()
	// testSendTransaction()
	// NewKey()
	// DecryptoJSON(utils.KeyDir + "UTC--2019-07-14T06-50-46.730758800Z--7a0934dc4ab137224006423f53505a6a732c8969")
}
