// Copyright 2016 Google Inc.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package utils

import (
	"crypto/rand"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"path/filepath"
	"time"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
)

// UUID user id
type UUID []byte

var rander = rand.Reader // random function

func randomBits(b []byte) {
	if _, err := io.ReadFull(rander, b); err != nil {
		panic(err.Error()) // rand should never fail
	}
}

// NewRandom get rand uuid
func NewRandom() UUID {
	uuid := make([]byte, 16)
	randomBits([]byte(uuid))
	uuid[6] = (uuid[6] & 0x0f) | 0x40 // Version 4
	uuid[8] = (uuid[8] & 0x3f) | 0x80 // Variant is 10
	return uuid
}

// WriteKeyFile hdkeystor writed to file
func WriteKeyFile(file string, content []byte) error {
	// Create the keystore directory with appropriate permissions
	// in case it is not present yet.
	const dirPerm = 0700

	if err := os.MkdirAll(filepath.Dir(file), dirPerm); err != nil {
		return err
	}
	// Atomic write: create a temporary hidden file first
	// then move it into place. TempFile assigns mode 0600.
	f, err := ioutil.TempFile(filepath.Dir(file), "."+filepath.Base(file)+".tmp")
	if err != nil {
		return err
	}
	if _, err := f.Write(content); err != nil {
		f.Close()
		os.Remove(f.Name())
		return err
	}
	f.Close()
	// newFile := KeyFileName(file)
	// fmt.Println("newFile: ", newFile)
	return os.Rename(f.Name(), file)
}

// KeyFileName ...
func KeyFileName(keyAddr string) string {
	ts := time.Now().UTC()
	// return fmt.Sprintf("UTC--%s--%s", toISO8601(ts), hex.EncodeToString(keyAddr[:]))
	return fmt.Sprintf("UTC--%s--%s", toISO8601(ts), keyAddr)
}

// toISO8601 ...
func toISO8601(t time.Time) string {
	var tz string
	name, offset := t.Zone()
	if name == "UTC" {
		tz = "Z"
	} else {
		tz = fmt.Sprintf("%03d00", offset/3600)
	}
	return fmt.Sprintf("%04d-%02d-%02dT%02d-%02d-%02d.%09d%s",
		t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), tz)
}

// GetKey by address, filename, auth
func GetKey(addr common.Address, filename, auth string) (*keystore.Key, error) {
	// Load the key from the keystore and decrypt its contents
	// log.Printf("addr: %s, filename: %s, auth: %s \n", addr.String(), filename, auth)
	keyjson, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Println("failed to ioutil.ReadFile")
		return nil, err
	}
	key, err := keystore.DecryptKey(keyjson, auth)
	if err != nil {
		log.Println("failed to keystore.DecryptKey")
		return nil, err
	}
	// Make sure we're really operating on the requested key (no swap attacks)
	if key.Address != addr {
		return nil, fmt.Errorf("key content mismatch: have account %x, want %x", key.Address, addr)
	}
	fmt.Println("key.Address: ", key.Address)
	return key, nil
}

// WalletDir ...
func WalletDir(buffer, filename string) string {
	if filepath.IsAbs(filename) {
		return filename
	}
	return filepath.Join(buffer, filename)
}

// Hex2bigInt ...
func Hex2bigInt(hex string) *big.Int {
	n := new(big.Int)
	n, _ = n.SetString(hex[2:], 16)
	return n
}
