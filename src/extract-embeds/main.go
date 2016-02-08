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

type Embed struct {
	Tag string
	Type string
}

func extractEmbeds(content string) ([]Embed, error) {
	tokenizer := html.NewTokenizer(bytes.NewBufferString(content))
	embeds := []Embed{}
	for {
		tt := tokenizer.Next()
		switch tt {
		case html.ErrorToken:
			return embeds, nil
		case html.SelfClosingTagToken:
			fallthrough
		case html.StartTagToken:
			tok := tokenizer.Token()
			if tok.Data == "embed" || tok.Data == "EMBED" {
				embedType := ""
				for _, attr := range tok.Attr {
					if attr.Key == "type" {
						embedType = attr.Val
					}
				}
				if embedType != "" {
					embeds = append(embeds, Embed{
						Tag: tok.Data,
						Type: embedType,
					})
				}
			}
		}
	}
	return embeds, nil
}

func main() {
	uris := []string{}
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		uris = append(uris, scanner.Text())
	}

	cache := pagecache.Cache{Dir: ".pagecache"}

	csvWriter := csv.NewWriter(os.Stdout)
	csvWriter.Write([]string{"Annotated URI", "Embed Type"})

	for _, uri := range uris {
		content, err := cache.Read(uri)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Skipping %s: %v\n", uri, err)
			continue
		}

		embeds, err := extractEmbeds(string(content))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to extract embeds from %s\n", uri, err)
			continue
		}
		for _, embed := range embeds {
			csvWriter.Write([]string{uri, embed.Type})
		}
	}
	csvWriter.Flush()
}

