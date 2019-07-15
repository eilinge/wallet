# issue

## 钱包

    验证keystore的正确性
        passphrase: KeyEncryptDecrypt()

## 交易

    {
    "address": "45dea0fb0bba44f4fcf290bba71fd57d7117cbb8",
    "crypto": {
        "cipher": "aes-128-ctr",
        "ciphertext": "b87781948a1befd247bff51ef4063f716cf6c2d3481163e9a8f42e1f9bb74145",
        "cipherparams": {
            "iv": "dc4926b48a105133d2f16b96833abf1e"
        },
        "kdf": "scrypt",
        "kdfparams": {
            "dklen": 32,
            "n": 2,
            "p": 1,
            "r": 8,
            "salt": "004244bbdc51cadda545b1cfa43cff9ed2ae88e08c61f1479dbb45410722f8f0"
        },
        "mac": "39990c1684557447940d4c69e06b1b82b2aceacb43f284df65c956daf3046b85"
    },
    "id": "ce541d8d-c79b-40f8-9f8c-20f59616faba",
    "version": 3
    }

## 文件结构

    keyStore
        key: 生成私钥
        passphrase: 生成keyStore; 验证keyStore
        keystore: 存储文件

## 钱包地址

    1. 账户名: (已获得)
            account.Address.Hex() // 0xb2E4BEec903EDB94054b4f91C1722A691F82a6C6
    2. 生成相应文件
        func keyFileName(keyAddr common.Address) string {
            ts := time.Now().UTC()
            return fmt.Sprintf("UTC--%s--%s", toISO8601(ts), hex.EncodeToString(keyAddr[:]))
            }
    3. 构成钱包数据
        newStore
            address:
            uuid:
            version:
        testMnemonic()
            cryptoJSON:
    4. 写入文件
        1. 遇到的问题, keystore中关于存储的函数为私有
            重写写入方法

## keystrore 实现逻辑

    1. passphrase.go(keyStorePassphrase) 实现了3个方法
        GetKey()
        StoreKey()
        JoinPath()
    
    2. keystore.go(keyStore) 实现了keyStorePassphrase3个方法的接口

## Key

    type keyStore interface {
        // Loads and decrypts the key from disk.
        GetKey(addr common.Address, filename string, auth string) (*Key, error)
        // Writes and encrypts the key.
        StoreKey(filename string, k *Key, auth string) error
        // Joins filename with the key directory unless it is already absolute.
        JoinPath(filename string) string
    }

    type plainKeyJSON struct {
        Address    string `json:"address"`
        PrivateKey string `json:"privatekey"`
        Id         string `json:"id"`
        Version    int    `json:"version"`
    }

    type encryptedKeyJSONV3 struct {
        Address string     `json:"address"`
        Crypto  CryptoJSON `json:"crypto"`
        Id      string     `json:"id"`
        Version int        `json:"version"`
    }

    type encryptedKeyJSONV1 struct {
        Address string     `json:"address"`
        Crypto  CryptoJSON `json:"crypto"`
        Id      string     `json:"id"`
        Version string     `json:"version"`
    }

    type CryptoJSON struct {
        Cipher       string                 `json:"cipher"`
        CipherText   string                 `json:"ciphertext"`
        CipherParams cipherparamsJSON       `json:"cipherparams"`
        KDF          string                 `json:"kdf"`
        KDFParams    map[string]interface{} `json:"kdfparams"`
        MAC          string                 `json:"mac"`
    }

    type cipherparamsJSON struct {
        IV string `json:"iv"`
    }

    key := &Key{
        Id:         id,
        Address:    crypto.PubkeyToAddress(privateKeyECDSA.PublicKey),
        PrivateKey: privateKeyECDSA,
    }

    UnmarshalJSON(cj []byte)

        u := new(uuid.UUID)
        *u = uuid.Parse(keyJSON.Id)
        k.Id = *u
        addr, err := hex.DecodeString(keyJSON.Address)
        privkey, err := crypto.HexToECDSA(keyJSON.PrivateKey)
        k.Address = common.BytesToAddress(addr)
        k.PrivateKey = privkey

    NewKeyForDirectICAP(rand io.Reader) *Key

    writeKeyFile(file string, content []byte)

    keyFileName(keyAddr common.Address) string

    toISO8601(t time.Time) string

    storeNewKey(ks keyStore, rand io.Reader, auth string) (*Key, accounts.Account, error) 

## keystore

    type KeyStore struct {
        storage  keyStore                     // Storage backend, might be cleartext or encrypted
        cache    *accountCache                // In-memory account cache over the filesystem storage
        changes  chan struct{}                // Channel receiving change notifications from the cache
        unlocked map[common.Address]*unlocked // Currently unlocked account (decrypted private keys)

        wallets     []accounts.Wallet       // Wallet wrappers around the individual key files
        updateFeed  event.Feed              // Event feed to notify wallet additions/removals
        updateScope event.SubscriptionScope // Subscription scope tracking current live listeners
        updating    bool                    // Whether the event notification loop is running

        mu sync.RWMutex
    }

    type unlocked struct {
        *Key
        abort chan struct{}
    }
