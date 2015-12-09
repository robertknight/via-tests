package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"pagecache"
	"strings"
)

func fetch(uri string) ([]byte, error) {
	resp, err := http.Get(uri)
	if err != nil {
		return nil, err
	}
	contentTypes := resp.Header["Content-Type"]
	if len(contentTypes) < 1 {
		return nil, fmt.Errorf("Unknown content type for URL %s", uri)
	}
	parts := strings.Split(contentTypes[0], ";")
	contentType := strings.TrimSpace(parts[0])
	if contentType != "text/html" {
		return nil, fmt.Errorf("Unsupported content type \"%s\"", contentType)
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
