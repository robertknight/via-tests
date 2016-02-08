package main

import (
	"bytes"
	"bufio"
	"encoding/csv"
	"fmt"
	"golang.org/x/net/html"
	"os"

	"pagecache"
)

type Link struct {
	Tag string
	Rel string
	Href string
	Type string
}

func extractLinks(content string) ([]Link, error) {
	tokenizer := html.NewTokenizer(bytes.NewBufferString(content))
	links := []Link{}
	for {
		tt := tokenizer.Next()
		switch tt {
		case html.ErrorToken:
			return links, nil
		case html.SelfClosingTagToken:
			fallthrough
		case html.StartTagToken:
			tok := tokenizer.Token()
			linkHref := ""
			linkRel := ""
			linkType := ""
			for _, attr := range tok.Attr {
				if attr.Key == "src" || attr.Key == "href" {
					linkHref = attr.Val
				} else if attr.Key == "rel" {
					linkRel = attr.Val
				} else if attr.Key == "type" {
					linkType = attr.Val
				}
			}
			if linkHref != "" {
				links = append(links, Link{
					Tag: tok.Data,
					Href: linkHref,
					Rel: linkRel,
					Type: linkType,
				})
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
	csvWriter.Write([]string{
		"Annotated URI",
		"Link Tag",
		"Link Rel",
		"Type",
		"Dest",
	})

	for _, uri := range uris {
		content, err := cache.Read(uri)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Skipping %s: %v\n", uri, err)
			continue
		}

		links, err := extractLinks(string(content))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to extract links from %s\n", uri, err)
			continue
		}
		for _, link := range links {
			linkDest := link.Href
			if link.Tag == "link" &&
			   link.Rel == "canonical" &&
			   len(linkDest) > 2 && linkDest[0] == '/' && linkDest[1] != '/' {
				// The ordering of fields here needs to be kept in sync
				// with the CSV headers (see above)
				csvWriter.Write([]string{
					uri,
					link.Tag,
					link.Rel,
					link.Type,
					linkDest,
				})
			}
		}
	}
	csvWriter.Flush()
}

