package client

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"strconv"
	"wallet/abi"
	"wallet/hdkeystore"
	"wallet/hdwallet"
	"wallet/utils"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
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
		TokensFile: "tokens.json",
	}
}

// Usage ...
func (cli *CLI) Usage() {
	fmt.Println("./wallet createwallet -name ACCOUNT_NAME -- for create a new wallet")
	fmt.Println("./wallet balance -addr ACCOUNT_NAME -- for get ether balance of a address")
	fmt.Println("./wallet transfer -from ACCOUNT_NAME -to ADDRESS -value VALUE -- for send ether to ADDRESS")
	fmt.Println("./wallet addtoken -addr CONTRACT_ADDR -- for send ether to ADDRESS")
	fmt.Println("./wallet tokenbalance -addr ACCOUNT_NAME -symbol TOKEN -- for get token balances")
	fmt.Println("./wallet sendtoken -from ACCOUNT_NAME -symbol SYMBOL -to ADDRESS -value VALUE -- for send tokens to ADDRESS")
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
	balancecmdAcct := balancecmd.String("addr", "", "ACCOUNT_NAME")

	// transfer -from address -to ADDRESS -value VALUE -- for send ether to ADDRESS
	transfer := flag.NewFlagSet("transfer", flag.ExitOnError)
	transferFrom := transfer.String("from", "", "from_Address")
	transferTo := transfer.String("to", "", "to_Address")
	transferValue := transfer.Int64("value", 0, "Value")

	// addtoken -addr CONTRACT_ADDR
	addtoken := flag.NewFlagSet("addtoken", flag.ExitOnError)
	tokenAddr := addtoken.String("addr", "", "Contact_Address")

	// tokenbalance -addr ACCOUNT_NAME -- for get token balances
	token := flag.NewFlagSet("tokenbalance", flag.ExitOnError)
	addr := token.String("addr", "", "Contact_Address")
	symbol := token.String("symbol", "", "TOKEN_SYMBOL")

	// sendtoken -addr ACCOUNT_NAME -symbol SYMBOL -toaddress ADDRESS -value VALUE
	sendtoken := flag.NewFlagSet("sendtoken", flag.ExitOnError)
	fromAddr := sendtoken.String("from", "", "Contact_Address")
	sendSymbol := sendtoken.String("symbol", "", "TOKEN_SYMBOL")
	toAddr := sendtoken.String("to", "", "Contact_Address")
	tokenValue := sendtoken.Int64("value", 0, "TOKEN_VALUE")

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
	case "transfer":
		err := transfer.Parse(os.Args[2:])

		if err != nil {
			log.Panic("failed to Parse balance params:", err)
		}

	case "addtoken":
		err := addtoken.Parse(os.Args[2:])

		if err != nil {
			log.Panic("failed to Parse addtoken params:", err)
		}

	case "tokenbalance":
		err := token.Parse(os.Args[2:])

		if err != nil {
			log.Panic("failed to Parse tokenbalance params:", err)
		}

	case "sendtoken":
		err := sendtoken.Parse(os.Args[2:])

		if err != nil {
			log.Panic("failed to Parse sendtoken params:", err)
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
			log.Fatal("transfer parames failed")
		}

		cli.Transfer(*transferFrom, *transferTo, *transferValue)
	}

	if addtoken.Parsed() {
		if *tokenAddr == "" {
			log.Fatal("addtoken parames failed")
		}

		cli.Addtoken(*tokenAddr)
	}

	// tokenbalance
	if token.Parsed() {
		if *addr == "" || *symbol == "" {
			log.Fatal("tokenbalance parames failed")
		}

		cli.TokenBalance(*addr, *symbol)
	}

	// sendtoken
	if sendtoken.Parsed() {
		if *fromAddr == "" || *sendSymbol == "" || *toAddr == "" || *tokenValue == 0 {
			log.Fatal("sendtoken parames failed")
		}

		cli.SendToken(*fromAddr, *sendSymbol, *toAddr, *tokenValue)
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

// GetBalance ...
func (cli *CLI) GetBalance(account string) {
	rclient, err := cli.GetAccount(account)
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
	return utils.Hex2bigInt(result)
}

// Transfer auth, Key
func (cli *CLI) Transfer(from, to string, value int64) {
	fileName, _, _, account, err := cli.getAccountKey(from)
	client, err := ethclient.Dial(cli.NetworkURL)
	if err != nil {
		log.Panic("failed to Transfer when Dial ", err)
	}
	defer client.Close()
	// 获取当前nonce值
	nonce, err := client.NonceAt(context.Background(), common.HexToAddress(account), nil)
	if err != nil {
		log.Panic("failed to Transfer when NonceAt ", err)
	}
	fmt.Println(nonce)

	gasLimit := uint64(30000)
	gasPrice := big.NewInt(1000000)
	tx := types.NewTransaction(nonce, common.HexToAddress(to), big.NewInt(value), gasLimit, gasPrice, []byte("salary"))

	hdks := hdkeystore.NewHDKeyStore(cli.DataPath, nil)

	fmt.Println("Please input your password for transfer")
	auth, err := gopass.GetPasswd()
	_, err = hdks.GetKey(common.HexToAddress(account), fileName, string(auth))
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

	log.Printf("from: %s Transfer to: %s value: %d success\n", from, to, value)
}

// Addtoken ...
func (cli *CLI) Addtoken(contactAddr string) (err error) {
	data, tokens, err := cli.addtokenjson(contactAddr)
	if err != nil {
		log.Println("failed to cli.addtokenjson", err)
		return err
	}

	symbol, err := cli.GetSymbol(contactAddr)
	if err != nil {
		log.Println("failed to cli.GetSymbol", err)
		return err
	}

	newtoken := TokenConfig{symbol, contactAddr}
	tokens = append(tokens, newtoken)
	data, _ = json.Marshal(tokens)
	utils.WriteKeyFile("tokens.json", data)
	log.Println("add token successfully")
	return err
}

func (cli *CLI) addtokenjson(contactAddr string) (data []byte, tokens []TokenConfig, err error) {
	tokens = []TokenConfig{}

	data, err = ioutil.ReadFile(cli.TokensFile)
	if err != nil {
		log.Fatal("failed to ioutil.ReadFile")
	}
	err = json.Unmarshal(data, &tokens)
	if err != nil {
		log.Fatal("failed to json.Unmarshal")
	}
	fmt.Println(tokens)

	if ok, _ := checkUniqueByAddr(contactAddr, tokens); !ok {
		log.Fatal("failed to checkUnique ")
	}
	return data, tokens, nil
}

func checkUniqueByAddr(address string, tokens []TokenConfig) (bool, error) {
	for _, token := range tokens {
		if token.Addr == address {
			return false, fmt.Errorf("token: %s existed", token.Symbol)
		}
	}
	return true, nil
}

func checkUniqueBySymbol(symbol string, tokens []TokenConfig) (bool, error) {
	for _, token := range tokens {
		if token.Symbol == symbol {
			return false, fmt.Errorf("token: %s existed", token.Symbol)
		}
	}
	return true, nil
}

// GetSymbol ...
func (cli *CLI) GetSymbol(address string) (string, error) {
	client, err := ethclient.Dial(cli.NetworkURL)
	if err != nil {
		log.Panic("failed to GetSymbol when Dial:", err)
	}
	defer client.Close()
	pxc, err := abi.NewPxc(common.HexToAddress(address), client)
	if err != nil {
		log.Panic("failed to abi.NewPxc:", err)
	}
	return pxc.Symbol(nil)
}

// TokenBalance ...
func (cli *CLI) TokenBalance(addr, symbol string) (int64, error) {
	tokenAddr, err := cli.getSymbolAddr(symbol)
	if err != nil {
		log.Panicln("failed to cli.getSymbolAddr: ", err)
	}
	client, err := ethclient.Dial(cli.NetworkURL)
	if err != nil {
		log.Panic("failed to GetSymbol when Dial:", err)
	}
	defer client.Close()
	pxc, err := abi.NewPxc(common.HexToAddress(tokenAddr), client)
	if err != nil {
		log.Panic("failed to abi.NewPxc:", err)
	}
	balance, err := pxc.BalanceOf(nil, common.HexToAddress(addr))
	if err != nil {
		log.Panic("failed to pxc.BalanceOf:", err)
	}
	fmt.Printf("your symbol: %s balance is: %d \n", symbol, balance.Int64())
	return balance.Int64(), nil
}

// SendToken ...
func (cli *CLI) SendToken(from, symbol, to string, value int64) {
	tokenAddr, err := cli.getSymbolAddr(symbol)
	if err != nil {
		log.Panicln("failed to cli.ggetSymbolAddr: ", err)
	}
	fileName, _, _, _, err := cli.getAccountKey(from)

	fmt.Println("get your filename: ", fileName)
	opt, err := cli.makeAuth(from, fileName)
	if err != nil {
		log.Panicln("failed to cli.makeAuth: ", err)
	}

	pxc, err := cli.getContact(tokenAddr)
	if err != nil {
		log.Panicln("failed to cli.getContact: ", err)
	}

	txhash, err := pxc.Transfer(opt, common.HexToAddress(to), big.NewInt(value))
	if err != nil {
		log.Panic("failed to Transfer ", err)
	}
	fmt.Println("sendtoken call ok,hash=", txhash.Hash().Hex())
}

func (cli *CLI) getSymbolAddr(symbol string) (string, error) {
	tokens := []TokenConfig{}
	data, err := ioutil.ReadFile(cli.TokensFile)
	if err != nil {
		log.Fatal("failed to ioutil.ReadFile")
	}
	err = json.Unmarshal(data, &tokens)
	if err != nil {
		log.Fatal("failed to json.Unmarshal")
	}

	fmt.Println("symbol: ", symbol)
	if ok, _ := checkUniqueBySymbol(symbol, tokens); ok {
		log.Fatal("the token symbol is not exist ")
	}

	var tokenAddr string
	for _, v := range tokens {
		if symbol == v.Symbol {
			tokenAddr = v.Addr
		}
	}
	return tokenAddr, nil
}

func (cli *CLI) makeAuth(from, fileName string) (opt *bind.TransactOpts, err error) {
	hdks := hdkeystore.NewHDKeyStore(cli.DataPath, nil)
	fmt.Println("Please input your password for transfer")
	auth, err := gopass.GetPasswd()
	_, err = hdks.GetKey(common.HexToAddress(from), fileName, string(auth))
	if err != nil {
		log.Panic("failed to Transfer when GetKey ", err)
	}
	keyin, err := os.Open(fileName)
	if err != nil {
		log.Panic("failed to read keystore file ", err)
	}
	opt, err = bind.NewTransactor(keyin, string(auth))
	if err != nil {
		log.Panic("failed to bind.NewTransactor ", err)
	}
	return
}

func (cli *CLI) getContact(tokenAddr string) (pxc *abi.Pxc, err error) {
	client, err := ethclient.Dial(cli.NetworkURL)
	if err != nil {
		log.Panic("failed to GetSymbol when Dial:", err)
	}
	defer client.Close()
	pxc, err = abi.NewPxc(common.HexToAddress(tokenAddr), client)
	return
}

// GetAccount ...
func (cli *CLI) GetAccount(account string) (rclient *rpc.Client, err error) {
	rclient, err = rpc.Dial(cli.NetworkURL)
	if err != nil {
		log.Fatal("failed to rpc.Dial ...")
	}
	return rclient, err
}

// getAccountKey ...
func (cli *CLI) getAccountKey(account string) (fileName string, rclient *rpc.Client, key *keystore.Key, accountAddr string, err error) {
	rclient, err = rpc.Dial(cli.NetworkURL)
	if err != nil {
		log.Panic("failed to Dial:", err)
	}
	defer rclient.Close()

	fmt.Println("Please input your wallet")

	var buffer string
	for {
		inputReader := bufio.NewReader(os.Stdin)
		buffer, err = inputReader.ReadString('\n')
		buffer = buffer[:len(buffer)-2]
		if err == nil && buffer != "" {
			break
		}
		fmt.Println("Input walletDir err or nil.Please input agin")
	}

	infos, err := ioutil.ReadDir(cli.DataPath + "/" + buffer)
	if err != nil {
		fmt.Println("failed to ReadDir:", err)
		return
	}
	fmt.Println("Please input your password for get key")
	auth, err := gopass.GetPasswd()

	for _, info := range infos {

		key, err := utils.GetKey(common.HexToAddress(account), utils.WalletDir(cli.DataPath+"/"+buffer, info.Name()), string(auth))
		if err == nil {
			accountAddr = account
			fileName = utils.WalletDir(cli.DataPath+"/"+buffer, info.Name())
			return fileName, rclient, key, accountAddr, nil
		}
	}
	return "", nil, nil, "", fmt.Errorf("key content mismatch: have account%s", account)
}
