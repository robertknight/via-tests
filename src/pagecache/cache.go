package pagecache

import (
	"crypto/sha1"
	"fmt"
	"encoding/hex"
	"io"
	"io/ioutil"
	"os"
	"path"
)

func shasum(text string) string {
	h := sha1.New()
	io.WriteString(h, text)
	return hex.EncodeToString(h.Sum(nil))
}


type Cache struct {
	Dir string
}

func (c *Cache) filePath(key string) string {
	keyHash := shasum(key)
	return c.Dir + "/" + keyHash[0:2] + "/" + keyHash[2:]
}

func (c *Cache) Write(key string, data []byte) error {
	cacheFilePath := c.filePath(key)
	cacheFileDir := path.Dir(cacheFilePath)
	err := os.MkdirAll(cacheFileDir, 0775)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating dir %s: %v", cacheFileDir, err)
		return err
	}
	cacheFile, err := os.Create(cacheFilePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create cache file %s: %v", cacheFilePath, err)
		return err
	}

	_, err = cacheFile.Write(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write cache data %s: %v", cacheFilePath, err)
		return err
	}

	err = cacheFile.Close()
	return err
}

func (c *Cache) Has(key string) bool {
	if _, err := os.Stat(c.filePath(key)); err == nil {
		return true
	} else {
		return false
	}
}

func (c *Cache) Read(key string) ([]byte, error) {
	return ioutil.ReadFile(c.filePath(key))
}

