package client

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"strconv"
	"strings"
	"wallet/hdkeystore"
	"wallet/hdwallet"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/howeyc/gopass"
)

// CLI ...
type CLI struct {
	DataPath   string
	NetworkURL string
	TokensFile string
}

// TokenConfig ...
type TokenConfig struct {
	Symbol string `json:"symbol"`
	Addr   string `json:"addr"`
}

// NewCLI ...
func NewCLI(path, url string) *CLI {
	return &CLI{
		DataPath:   path,
		NetworkURL: url,
		TokensFile: "token.json",
	}
}

// Usage ...
func (cli *CLI) Usage() {
	fmt.Println("./wallet createwallet -name ACCOUNT_NAME -- for create a new wallet")
	fmt.Println("./wallet balance -name ACCOUNT_NAME -- for get ether balance of a address")
	fmt.Println("./wallet transfer -name ACCOUNT_NAME -toaddress ADDRESS -value VALUE -- for send ether to ADDRESS")
	fmt.Println("./wallet addtoken -addr CONTRACT_ADDR -- for send ether to ADDRESS")
	fmt.Println("./wallet tokenbalance -name ACCOUNT_NAME -- for get token balances")
	fmt.Println("./wallet sendtoken -name ACCOUNT_NAME -symbol SYMBOL -toaddress ADDRESS -value VALUE -- for send tokens to ADDRESS")
}

func (cli *CLI) validateArgs() {
	if len(os.Args) < 2 {
		cli.Usage()
		os.Exit(1)
	}
}

// Run ...
func (cli *CLI) Run() {
	cli.validateArgs()

	// 绑定
	createwalletcmd := flag.NewFlagSet("createwallet", flag.ExitOnError)
	// 假使不使用 createwallet -name eilinge, 则会传递该默认值
	createwalletcmdAcct := createwalletcmd.String("name", "tester", "ACCOUNT_NAME")

	balancecmd := flag.NewFlagSet("balance", flag.ExitOnError)
	balancecmdAcct := balancecmd.String("account", "", "ACCOUNT_NAME")

	// transfer -from address -to ADDRESS -value VALUE -- for send ether to ADDRESS
	transfer := flag.NewFlagSet("transfer", flag.ExitOnError)
	transferFrom := transfer.String("from", "", "from_Address")
	transferTo := transfer.String("to", "", "to_Address")
	transferValue := transfer.Int64("value", 0, "Value")

	switch os.Args[1] {
	case "createwallet":
		// 获取
		err := createwalletcmd.Parse(os.Args[2:])

		if err != nil {
			log.Panic("failed to Parse createwallet params:", err)
		}
	// balance -name ACCOUNT_NAME -- for get ether balance of a address"
	case "balance":
		err := balancecmd.Parse(os.Args[2:])

		if err != nil {
			log.Panic("failed to Parse balance params:", err)
		}
	case "tranfer":
		err := transfer.Parse(os.Args[2:])

		if err != nil {
			log.Panic("failed to Parse balance params:", err)
		}
	default:
		cli.Usage()
		os.Exit(1)
	}

	// 解析
	if createwalletcmd.Parsed() {
		fmt.Println("*createwalletcmdAcct: ", *createwalletcmdAcct)
		if !cli.checkPath(*createwalletcmdAcct) {
			fmt.Println("the keystore director is not null,you can not create wallet!")
			os.Exit(1)
		}
		fmt.Println("call create wallet, Please input your password for keystore")
		// 隐藏密码
		pass, err := gopass.GetPasswd()

		if err != nil {
			log.Panic("failed to get your password:", err)
		}

		cli.CreateWallet(*createwalletcmdAcct, string(pass))

		log.Println("CreateWallet success ...")
	}

	if balancecmd.Parsed() {
		if *balancecmdAcct != "" {
			cli.GetBalance(*balancecmdAcct)
		}

	}

	if transfer.Parsed() {
		if *transferFrom == "" || *transferTo == "" || *transferValue == 0 {
			log.Fatal("transfer parames")
		}

		cli.Transfer(*transferFrom, *transferTo, *transferValue)
	}
}

func (cli *CLI) checkPath(name string) bool {
	infos, err := ioutil.ReadDir(cli.DataPath + "/" + name)
	if err != nil {
		// fmt.Println("failed to ReadDir:", err)
		log.Println("you can build the filename for store account")
		return true
	}
	if len(infos) > 0 {
		return false
	}
	return true
}

// CreateWallet ...
func (cli *CLI) CreateWallet(name, pass string) {
	mnemonic, err := hdwallet.NewMnemonic(160)
	if err != nil {
		log.Panic("failed to NewMnemonic:", err)
	}

	fmt.Printf("Please remember the mnemonic:\n[%s]\n\n", mnemonic)

	wallet, err := hdwallet.NewFromMnemonic(mnemonic, "")
	if err != nil {
		log.Panic("failed to NewFromMnemonic:", err)
	}

	for i := 0; i < 1; i++ {
		path := hdwallet.MustParseDerivationPath("m/44'/60'/0'/0/" + strconv.Itoa(i))
		// "m/44'/60'/0'/0/0" -> common.Address
		account, err := wallet.Derive(path, true)
		if err != nil {
			log.Panic("failed to Derive:", err)
		}
		fmt.Printf("%s account.Address: %s ", strconv.Itoa(i), account.Address.Hex())

		// common.Address -> pkey
		pkey, err := wallet.PrivateKey(account)
		if err != nil {
			log.Panic("failed to PrivateKey:", err)
		}

		hdks := hdkeystore.NewHDKeyStore(cli.DataPath+"/"+name, pkey)
		// hdks -> UTC-address
		err = hdks.StoreKey(account.Address.Hex(), pass)
		if err != nil {
			log.Panic("failed to store key:", err)
		}
	}
}

// GetBalance ...
func (cli *CLI) GetBalance(account string) {
	_, rclient, _, account, err := cli.getAccount(account)
	if err != nil {
		log.Fatal("failed get account file")
	}
	fmt.Printf("The balance of %s is %v\n", account, cli.getBalance(account, rclient))
}

func (cli *CLI) getBalance(address string, client *rpc.Client) *big.Int {
	var result string
	err := client.Call(&result, "eth_getBalance", common.HexToAddress(address), "latest")
	if err != nil {
		log.Panic("failed to call eth_getBalance:", err)
	}
	return hex2bigInt(result)
}

func hex2bigInt(hex string) *big.Int {
	n := new(big.Int)
	n, _ = n.SetString(hex[2:], 16)
	return n
}

// Transfer auth, Key
func (cli *CLI) Transfer(from, to string, value int64) {
	fileName, _, _, account, err := cli.getAccount(from)
	client, err := ethclient.Dial(cli.NetworkURL)
	if err != nil {
		log.Panic("failed to Transfer when Dial ", err)
	}
	defer client.Close()
	//获取当前nonce值
	nonce, err := client.NonceAt(context.Background(), common.HexToAddress(account), nil)
	if err != nil {
		log.Panic("failed to Transfer when NonceAt ", err)
	}
	fmt.Println(nonce)

	gasLimit := uint64(300000)
	gasPrice := big.NewInt(21000000000)
	tx := types.NewTransaction(nonce, common.HexToAddress(to), big.NewInt(value), gasLimit, gasPrice, []byte("salary"))

	hdks := hdkeystore.NewHDKeyStore(cli.DataPath, nil)

	auth, err := gopass.GetPasswd()
	hdks.GetKey(common.HexToAddress(account), hdks.JoinPath(fileName), string(auth))
	if err != nil {
		log.Panic("failed to Transfer when GetKey ", err)
	}

	stx, err := hdks.SignTx(common.HexToAddress(from), tx, nil)
	if err != nil {
		log.Panic("failed to Transfer when SignTx ", err)
	}

	err = client.SendTransaction(context.Background(), stx)
	if err != nil {
		log.Panic("failed to Transfer when SendTransaction ", err)
	}

}

// getAccount ...
func (cli *CLI) getAccount(account string) (fileName string, rclient *rpc.Client, key *keystore.Key, accountAddr string, err error) {
	rclient, err = rpc.Dial(cli.NetworkURL)
	if err != nil {
		log.Panic("failed to Dial:", err)
	}
	defer rclient.Close()
	infos, err := ioutil.ReadDir(cli.DataPath + "/keystore")
	if err != nil {
		fmt.Println("failed to ReadDir:", err)
		return
	}
	for _, info := range infos {
		if strings.Index(info.Name(), account) != -1 {
			hdks := hdkeystore.NewHDKeyStore(cli.DataPath, nil)

			fmt.Println("info.Name(): ", hdks.JoinPath(info.Name()))

			keyjson, err := ioutil.ReadFile(hdks.JoinPath(info.Name()))
			if err != nil {
				log.Fatal("failed to ioutil.ReadFile ...")
			}
			auth, err := gopass.GetPasswd()

			key, err := keystore.DecryptKey(keyjson, string(auth))
			if err != nil {
				log.Fatal("failed to keystore.DecryptKey ...")
			}
			// Make sure we're really operating on the requested key (no swap attacks)
			if key.Address != common.HexToAddress(account) {
				log.Fatal("key content mismatch")
			}
			accountAddr = account
			fileName = info.Name()
			return fileName, rclient, key, accountAddr, nil
		}
	}
	return "", nil, nil, "", fmt.Errorf("key content mismatch: have account%s", account)
}
