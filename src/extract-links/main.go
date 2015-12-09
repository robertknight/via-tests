package main

import (
	"bytes"
	"bufio"
	"encoding/csv"
	"fmt"
	"golang.org/x/net/html"
	"os"
	"net/url"

	"pagecache"

	"strconv"
)

type Link struct {
	Tag string
	Href string
}

type PageLinks struct {
	URL string
	Links []Link
}

func extractLinks(content string) ([]Link, error) {
	tokenizer := html.NewTokenizer(bytes.NewBufferString(content))
	links := []Link{}
	for {
		tt := tokenizer.Next()
		switch tt {
		case html.ErrorToken:
			return links, nil
		case html.StartTagToken:
			tok := tokenizer.Token()
			for _, attr := range tok.Attr {
				if attr.Key == "src" || attr.Key == "href" {
					links = append(links, Link{
						Tag: tok.Data,
						Href: attr.Val,
					})
				}
			}
		}
	}
	return links, nil
}

func main() {
	uris := []string{}
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		uris = append(uris, scanner.Text())
	}

	cache := pagecache.Cache{Dir: ".pagecache"}

	csvWriter := csv.NewWriter(os.Stdout)
	csvWriter.Write([]string{"Annotated URI", "Total Links", "Absolute", "Absolute (Same Host)", "Relative"})

	for _, uri := range uris {
		content, err := cache.Read(uri)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Skipping %s: %v\n", uri, err)
			continue
		}

		parsedUri, err := url.Parse(uri)
		if err != nil {
			continue
		}

		absUrls := 0
		relUrls := 0
		sameHostAbsUrls := 0

		links, err := extractLinks(string(content))
		for _, link := range links {
			url, err := url.Parse(link.Href)
			if err != nil {
				continue
			}
			if url.IsAbs() {
				absUrls += 1
				if url.Host == parsedUri.Host {
					sameHostAbsUrls += 1
				}
			} else {
				relUrls += 1
			}
		}

		totalUris := absUrls + relUrls
		err = csvWriter.Write([]string{
			uri,
			strconv.Itoa(totalUris),
			strconv.Itoa(absUrls),
			strconv.Itoa(sameHostAbsUrls),
			strconv.Itoa(relUrls),
		})
		csvWriter.Flush()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
		}
	}
}

