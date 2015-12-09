package main

import (
	"fmt"
	"os"
	"hypothesis"
)

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

func main() {
	maxUrls := 2000
	fmt.Fprintf(os.Stderr, "Fetching last %d annotated URLs\n", maxUrls)
	uris := fetchLastNAnnotatedUris(maxUrls)
	for _, uri := range uris {
		fmt.Println(uri)
	}
}

