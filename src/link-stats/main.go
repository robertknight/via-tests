package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"hypothesis"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"golang.org/x/net/html"
)

type link struct {
	href string
}

func fetchLastNAnnotatedUris(n int) []string {
	uniqueUris := map[string]int{}
	prevOffset := 0
	for len(uniqueUris) < n {
		fmt.Fprintf(os.Stderr, "Fetching from offset %v\n", prevOffset)
		q := hypothesis.SearchQuery{
			Offset: prevOffset,
			Limit:  500,
		}
		r, err := hypothesis.Search(q)
		if err != nil {
			fmt.Fprintf(os.Stderr, "API search failed: %v\n", err)
			continue
		}
		prevOffset += len(r.Rows)
		for _, annot := range r.Rows {
			if len(uniqueUris) >= n {
				break
			}
			uniqueUris[annot.Uri] += 1
		}
	}
	uris := []string{}
	for uri, _ := range uniqueUris {
		uris = append(uris, uri)
	}
	return uris
}

func fetch(uri string) (string, error) {
	resp, err := http.Get(uri)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("request failed with status %d", resp.StatusCode)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func extractLinks(content string) ([]link, error) {
	tokenizer := html.NewTokenizer(bytes.NewBufferString(content))
	links := []link{}
	for {
		tt := tokenizer.Next()
		switch tt {
		case html.ErrorToken:
			return links, nil
		case html.StartTagToken:
			tok := tokenizer.Token()
			if tok.Data == "a" {
				for _, attr := range tok.Attr {
					if attr.Key == "href" {
						links = append(links, link{href: attr.Val})
					}
				}
			}
		}
	}
	return links, nil
}

func main() {
	maxUrls := 200

	fmt.Fprintf(os.Stderr, "Fetching last %d annotated URLs\n", maxUrls)

	csvWriter := csv.NewWriter(os.Stdout)
	csvWriter.Write([]string{"Annotated URI", "Total Links", "Absolute", "Absolute (Same Host)", "Relative"})

	uris := fetchLastNAnnotatedUris(maxUrls)
	for _, uri := range uris {
		content, err := fetch(uri)
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

		links, err := extractLinks(content)
		for _, link := range links {
			url, err := url.Parse(link.href)
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
