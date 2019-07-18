package utils

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// KeyDir ...
var (
	KeyDir = "./data/keystore/"
)

// KeyFileName ...
func KeyFileName(keyAddr common.Address) string {
	ts := time.Now().UTC()
	return fmt.Sprintf("UTC--%s--%s", toISO8601(ts), hex.EncodeToString(keyAddr[:]))
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

// WriteKeyFile ...
func WriteKeyFile(file string, content []byte) error {
	KeyStore := KeyDir + file
	name, err := writeTemporaryKeyFile(KeyStore, content)
	if err != nil {
		return err
	}
	// windows 切换文件目录, 否则找不到相应文件
	os.Chdir(KeyDir)
	defer os.Chdir("../../")
	// fmt.Printf("old: %s, new:%s \n", name, file)
	return os.Rename(name, file)
}

// writeTemporaryKeyFile ...
func writeTemporaryKeyFile(file string, content []byte) (string, error) {
	// Create the keystore directory with appropriate permissions
	// in case it is not present yet.

	const dirPerm = 0700
	fileDir := filepath.Dir(file)
	if err := os.MkdirAll(fileDir, dirPerm); err != nil {
		return "", err
	}

	// Atomic write: create a temporary hidden file first
	// then move it into place. TempFile assigns mode 0600.
	f, err := ioutil.TempFile(filepath.Dir(file), filepath.Base(file)+".tmp")

	defer f.Close()
	if err != nil {
		return "", err
	}

	if err != nil {
		return "", err
	}

	if _, err := f.Write(content); err != nil {
		os.Remove(f.Name())
		return "", err
	}
	fmt.Println("filepath.Base(f.Name()): ", filepath.Base(f.Name()))
	return filepath.Base(f.Name()), nil
}
