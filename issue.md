# Issue

## geth

    geth --datadir ./data --networkid 15 --port 30303 --rpc --rpcaddr 0.0.0.0 --rpcport 8545 --rpcvhosts "*" --rpcapi "db,net,eth,web3,personal" --rpccorsdomain "*" --ws --wsaddr "localhost" --wsport "8546" --wsorigins "*" --nat "any" --nodiscover --dev --dev.period 1 console 2> 1.log

## mysqld server

## 带需解决的问题

    1. 未实现该钱包捕捉交易的事件
    2. hdwallet的其他函数有待验证
    3. HD钱包路径参数的含义
