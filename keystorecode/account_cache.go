// Copyright 2017 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package keystorecode

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	mapset "github.com/deckarep/golang-set"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
)

// Minimum amount of time between cache reloads. This limit applies if the platform does
// not support change notifications. It also applies if the keystore directory does not
// exist yet, the code will attempt to create a watcher at most this often.
const minReloadInterval = 2 * time.Second

// wallet中, 存储账户
/*
type Account struct {
	Address common.Address `json:"address"` // Ethereum account address derived from the key
	URL     URL            `json:"url"`     // keystore 文件名
}
*/
type accountsByURL []accounts.Account

func (s accountsByURL) Len() int           { return len(s) }
func (s accountsByURL) Less(i, j int) bool { return s[i].URL.Cmp(s[j].URL) < 0 }
func (s accountsByURL) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

// AmbiguousAddrError is returned when attempting to unlock
// an address for which more than one file exists.
type AmbiguousAddrError struct {
	Addr    common.Address
	Matches []accounts.Account
}

func (err *AmbiguousAddrError) Error() string {
	files := ""
	for i, a := range err.Matches {
		files += a.URL.Path
		if i < len(err.Matches)-1 {
			files += ", "
		}
	}
	return fmt.Sprintf("multiple keys match address (%s)", files)
}

// accountCache is a live index of all accounts in the keystore.
type accountCache struct {
	keydir   string // get addresses by keydir
	watcher  *watcher
	mu       sync.Mutex                            // lock when operate
	all      accountsByURL                         // []accounts.Account
	byAddr   map[common.Address][]accounts.Account // 一个地址(Address), 存储多个账户[]accounts.Account{Address, URL}
	throttle *time.Timer
	notify   chan struct{}
	fileC    fileCache
}

func newAccountCache(keydir string) (*accountCache, chan struct{}) {
	ac := &accountCache{
		keydir: keydir,
		byAddr: make(map[common.Address][]accounts.Account),
		notify: make(chan struct{}, 1), // 创建一个新账户时, 缓存1个
		fileC:  fileCache{all: mapset.NewThreadUnsafeSet()},
	}
	ac.watcher = newWatcher(ac) // 生成一个watcher
	return ac, ac.notify
}

func (ac *accountCache) accounts() []accounts.Account {
	ac.maybeReload()
	// 对accountCache 操作(accounts/hasAddress/add/delete/deleteByFile/maybeReload)
	// 时进行上锁, 防止其他人同时进行操作
	ac.mu.Lock()
	defer ac.mu.Unlock()
	// 返回该accountCache的[]accounts.Account
	cpy := make([]accounts.Account, len(ac.all))
	copy(cpy, ac.all)
	return cpy
}

func (ac *accountCache) hasAddress(addr common.Address) bool {
	ac.maybeReload()
	ac.mu.Lock()
	defer ac.mu.Unlock()
	return len(ac.byAddr[addr]) > 0
}

func (ac *accountCache) add(newAccount accounts.Account) {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	// 通过闭包, 判断该accountCache中是否已经存在该账户
	// x.Cmp(y): -1 if x < y 0 if x == y 1 if x > y
	// Search uses binary search to find and return the smallest index i in [0, n) at which f(i) is true
	// If there is no such index, Search returns n
	i := sort.Search(len(ac.all), func(i int) bool { return ac.all[i].URL.Cmp(newAccount.URL) >= 0 }) // ?
	// 已经存在该账户
	if i < len(ac.all) && ac.all[i] == newAccount {
		return
	}
	// newAccount is not in the cache.
	ac.all = append(ac.all, accounts.Account{})
	// 添加
	copy(ac.all[i+1:], ac.all[i:])
	// 移除
	// ac.all = append(ac.all[:i], ac.all[i+1:]...)
	// 添加新的账户
	ac.all[i] = newAccount
	ac.byAddr[newAccount.Address] = append(ac.byAddr[newAccount.Address], newAccount)
}

// note: removed needs to be unique here (i.e. both File and Address must be set).
func (ac *accountCache) delete(removed accounts.Account) {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	ac.all = removeAccount(ac.all, removed)
	// []accounts.Account
	if ba := removeAccount(ac.byAddr[removed.Address], removed); len(ba) == 0 {
		//  假使在accounts.Account中仅其有一个, 移除account之后, accounts.Account为空, 清除ac.byAddr[account.Address]
		delete(ac.byAddr, removed.Address)
	} else {
		ac.byAddr[removed.Address] = ba
	}
}

// deleteByFile removes an account referenced by the given path.
func (ac *accountCache) deleteByFile(path string) {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	i := sort.Search(len(ac.all), func(i int) bool { return ac.all[i].URL.Path >= path })

	if i < len(ac.all) && ac.all[i].URL.Path == path {
		removed := ac.all[i]
		ac.all = append(ac.all[:i], ac.all[i+1:]...)
		if ba := removeAccount(ac.byAddr[removed.Address], removed); len(ba) == 0 {
			delete(ac.byAddr, removed.Address)
		} else {
			ac.byAddr[removed.Address] = ba
		}
	}
}

func removeAccount(slice []accounts.Account, elem accounts.Account) []accounts.Account {
	for i := range slice {
		if slice[i] == elem {
			// 移除查找到的account
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}

// find returns the cached account for address if there is a unique match.
// The exact matching rules are explained by the documentation of accounts.Account.
// Callers must hold ac.mu.
func (ac *accountCache) find(a accounts.Account) (accounts.Account, error) {
	// Limit search to address candidates if possible.
	matches := ac.all
	// 1. 根据a.Address相对应的[]accounts.Address进行查找
	if (a.Address != common.Address{}) {
		matches = ac.byAddr[a.Address]
	}
	// a.URL.Path
	if a.URL.Path != "" {
		// If only the basename is specified, complete the path.

		// 2. 根据路径查找
		// ContainsRune 判断字符串 s 中是否包含字符 r
		// os.PathSeparator: "/"
		if !strings.ContainsRune(a.URL.Path, filepath.Separator) {
			// s := filepath.Join("a", "b")
			// fmt.Println(s) // a/b/
			a.URL.Path = filepath.Join(ac.keydir, a.URL.Path)
		}
		for i := range matches {
			if matches[i].URL == a.URL {
				return matches[i], nil
			}
		}
		// 没有找到, 返回空地址
		if (a.Address == common.Address{}) {
			return accounts.Account{}, ErrNoMatch
		}
	}
	/*
		// a.Address
		switch len(matches) {
		// ac.byAddr[a.Address], 有且仅有一个, 已进行添加ac.byAddr[a.Address] = []ac.account
		// TODO: len(ac.byAddr[a.Address]) > 1
		// case len(matches) == 1:
		case 1:
			return matches[0], nil
		// ac.byAddr[a.Address] 不包含任何地址, 还未添加任何a.Address
		case 0:
			return accounts.Account{}, ErrNoMatch
		default:
			err := &AmbiguousAddrError{Addr: a.Address, Matches: make([]accounts.Account, len(matches))}
			copy(err.Matches, matches)
			sort.Sort(accountsByURL(err.Matches))
			return accounts.Account{}, err
		}
	*/
	for _, k := range matches {
		if k.Address == a.Address {
			return k, nil
		}
	}
	err := &AmbiguousAddrError{Addr: a.Address, Matches: make([]accounts.Account, len(matches))}
	copy(err.Matches, matches)
	sort.Sort(accountsByURL(err.Matches))
	return accounts.Account{}, err
}

func (ac *accountCache) maybeReload() {
	ac.mu.Lock()

	if ac.watcher.running {
		ac.mu.Unlock()
		return // A watcher is running and will keep the cache up-to-date.
	}
	if ac.throttle == nil {
		ac.throttle = time.NewTimer(0)
	} else {
		select {
		case <-ac.throttle.C:
		default:
			ac.mu.Unlock()
			return // The cache was reloaded recently.
		}
	}
	// No watcher running, start it.
	ac.watcher.start()
	// 重启计时器
	ac.throttle.Reset(minReloadInterval)
	ac.mu.Unlock()
	ac.scanAccounts()
}

func (ac *accountCache) close() {
	ac.mu.Lock()
	ac.watcher.close()
	// 关闭计时器
	if ac.throttle != nil {
		ac.throttle.Stop()
	}
	// 关闭ac.notify通道
	if ac.notify != nil {
		close(ac.notify)
		ac.notify = nil
	}
	ac.mu.Unlock()
}

// scanAccounts checks if any changes have occurred on the filesystem, and
// updates the account cache accordingly
func (ac *accountCache) scanAccounts() error {
	// Scan the entire folder metadata for file changes
	creates, deletes, updates, err := ac.fileC.scan(ac.keydir)
	if err != nil {
		log.Debug("Failed to reload keystore contents", "err", err)
		return err
	}
	// Returns the number of elements in the set.
	if creates.Cardinality() == 0 && deletes.Cardinality() == 0 && updates.Cardinality() == 0 {
		return nil
	}
	// Create a helper method to scan the contents of the key files
	var (
		buf = new(bufio.Reader)
		key struct {
			Address string `json:"address"`
		}
	)
	// 闭包函数:readAccount(path string), 解析keystore文件中的address
	readAccount := func(path string) *accounts.Account {
		fd, err := os.Open(path)
		if err != nil {
			log.Trace("Failed to open keystore file", "path", path, "err", err)
			return nil
		}
		defer fd.Close()
		buf.Reset(fd)
		// Parse the address.
		key.Address = ""
		err = json.NewDecoder(buf).Decode(&key)
		addr := common.HexToAddress(key.Address)
		switch {
		case err != nil:
			log.Debug("Failed to decode keystore key", "path", path, "err", err)
		case (addr == common.Address{}):
			log.Debug("Failed to decode keystore key", "path", path, "err", "missing or zero address")
		default:
			return &accounts.Account{
				Address: addr,
				URL:     accounts.URL{Scheme: KeyStoreScheme, Path: path},
			}
		}
		return nil
	}
	// Process all the file diffs
	start := time.Now()

	// set.Set -> slice
	for _, p := range creates.ToSlice() {
		// interface{}转换成string
		if a := readAccount(p.(string)); a != nil {
			ac.add(*a)
		}
	}
	for _, p := range deletes.ToSlice() {
		ac.deleteByFile(p.(string))
	}
	for _, p := range updates.ToSlice() {
		path := p.(string)
		ac.deleteByFile(path)
		if a := readAccount(path); a != nil {
			ac.add(*a)
		}
	}
	end := time.Now()

	select {
	// TODO: what is ac.notify role?
	case ac.notify <- struct{}{}:
	default:
	}
	log.Trace("Handled keystore changes", "time", end.Sub(start))
	return nil
}
