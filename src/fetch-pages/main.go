package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"pagecache"
)

func fetch(uri string) ([]byte, error) {
	resp, err := http.Get(uri)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("request failed with status %d", resp.StatusCode)
	}
	return ioutil.ReadAll(resp.Body)
}

func main() {
	cache := pagecache.Cache{Dir: ".pagecache"}
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		uri := scanner.Text()
		if cache.Has(uri) {
			continue
		}

		fmt.Fprintf(os.Stderr, "Fetching %s\n", uri)
		body, err := fetch(uri)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to fetch %s: %v\n", uri, err)
			continue
		}
		err = cache.Write(uri, body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to cache %s: %v\n", uri, err)
			continue
		}
	}
}
