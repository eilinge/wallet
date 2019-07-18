package client

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"wallet/hdkeystore"
	"wallet/hdwallet"

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
	createwalletcmdAcct := createwalletcmd.String("name", "tester", "ACCOUNT_NAME")

	switch os.Args[1] {
	case "createwallet":
		// 获取
		err := createwalletcmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic("failed to Parse createwallet params:", err)
		}
	default:
		cli.Usage()
		os.Exit(1)
	}

	// 解析
	if createwalletcmd.Parsed() {

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

	for i := 0; i < 10; i++ {
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
