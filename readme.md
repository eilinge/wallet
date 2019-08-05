# wallet

使用golang实现HDWallet钱包(https://www.jianshu.com/p/53405db83c16):

    1. 创建钱包: ./wallet.exe createwallet -name HDWALLET_NAME
    2. 查询ether余额: ./wallet.exe balance -addr ACCOUNT_ADDRSS
    3. 转账ether: ./wallet.exe transfer -from ACCOUNT_ADDRESS -to ADDRESS -value VALUE
    4. 添加token: ./wallet.exe addtoken -addr CONTRACT_ADDRSS
    5. 查询token余额: ./wallet.exe tokenbalance -addr ACCOUNT_ADDRESS -symbol TOKEN
    6. 转账token: ./wallet.exe sendtoken -from ACCOUNT_ADDRESS -symbol SYMBOL -to ADDRESS -value VALUE

## golang/geth 下载

    1. golang: go1.11.2 windows/amd64
    2. geth: 1.8.16-stable-477eb093
    3. wallet: git clone https://github.com/eilinge/wallet.git

## go依赖库

    1. github.com/ethereum/go-ethereum
    2. github.com/howeyc/gopass
    3. 其他依赖库, 请在编译wallet时, 按照提示进行使用"go get -u packageName"安装

## 操作流程

    1. 使用geth开启私链:
        `geth --datadir ./data --networkid 15 --port 30303 --rpc --rpcaddr 0.0.0.0 --rpcport 8545 --rpcvhosts "*" --rpcapi "db,net,eth,web3,personal" --rpccorsdomain "*" --ws --wsaddr "localhost" --wsport "8546" --wsorigins "*" --nat "any" --nodiscover --dev --dev.period 1 console 2> 1.log`
    
    2. 编译wallet
        `cd wall && go build -i`

    3. `./wallet.exe` 查看help

    4. 使用remix(http://remix.ethereum.org/#optimize=true&evmVersion=null&version=soljson-v0.4.26+commit.4563c3fc.js), 将solidity/ERC20.sol和pxcCoin.sol 部署至 http://localhost:8545

## 功能的使用

    1. 创建钱包: ./wallet.exe createwallet -name HDWALLET_NAME
        1. ./wallet.exe createwallet -name test
        2. 按照提示, 输入该钱包的密钥
        3. 会生成10个钱包地址文件存储至: data/test
    
    2. 查询ether余额: ./wallet.exe balance -addr ACCOUNT_ADDRSS
        1. 进入创建的钱包: cd data/test
        2. 通过keystore文件, 获取账户address: 0xF381BB62cD6695BbaE2f098B24AEF44CCD7b62c5
        3. ./wallet.exe balance -addr 0xF381BB62cD6695BbaE2f098B24AEF44CCD7b62c5

    3. 转账ether: ./wallet.exe transfer -from ACCOUNT_ADDRESS -to ADDRESS -value VALUE
        1. 通过geth, 向查询test/address转ether: `eth.sendTransaction({from:eth.accounts[0], to:"0xF381BB62cD6695BbaE2f098B24AEF44CCD7b62c5", value:10000000000})`
        2. ./wallet.exe transfer -from 0xD73f0ebC5f5BcE989138d8E8B05eA77d79f0D297 -to 0x9f24648A2c471f9ace923E788ff992729f2fAa7c -value 100
        3. 根据提示, 输入所要使用的钱包和创建钱包时的秘钥
        4. 成功的消息"2019/08/05 15:45:03 from: 0xD73f0ebC5f5BcE989138d8E8B05eA77d79f0D297 Transfer to: 0x9f24648A2c471f9ace923E788ff992729f2fAa7c value: 100 success"

    4. 添加token: ./wallet.exe addtoken -addr CONTRACT_ADDRSS
        1. 通过remix部署pxcCoin.sol 合约, 获取该合约地址"0x976486d0025bd4ebd905868f79c430d8f19c07cf"
        2. /wallet.exe addtoken -addr 0x976486d0025bd4ebd905868f79c430d8f19c07cf
        3. 添加成功: "2019/08/05 15:54:49 add token successfully"

    5. 查询token余额: ./wallet.exe tokenbalance -addr ACCOUNT_ADDRESS -symbol TOKEN
        1. 通过remix, 调用transfer函数, 向"0xD73f0ebC5f5BcE989138d8E8B05eA77d79f0D297"转入pxc token
        2. ./wallet.exe tokenbalance -addr 0xD73f0ebC5f5BcE989138d8E8B05eA77d79f0D297 -symbol pxc
        3. 查询余额成功: "your symbol: pxc balance is: 100000"

    6. 转账token: ./wallet.exe sendtoken -from ACCOUNT_ADDRESS -symbol SYMBOL -to ADDRESS -value
        1. ./wallet.exe sendtoken -from 0xD73f0ebC5f5BcE989138d8E8B05eA77d79f0D297 -symbol pxc -to 9f24648a2c471f9ace923e788ff992729f2faa7c -value 100
        2. 按照提示, 输入钱包文件和秘钥
        3. 转账成功信息: "sendtoken call ok,hash= 0xed82a4593d52beec2a3d0a4ee4402d16c0a30d6069f31a5df10a2172821e84a2"
