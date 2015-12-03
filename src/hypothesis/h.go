package hypothesis

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Annotation struct {
	Updated string
	Created string
	Tags []string
	Text string
	User string
	Id string
	Uri string

	// TODO
	// document
	// permissions
	// target
}

type SearchResult struct {
	Rows []Annotation
	Total int64
}

type SearchQuery struct {
	Offset int
	Limit int
}

func Search(query SearchQuery) (*SearchResult, error) {
	queryStr := "?"
	if query.Offset > 0 {
		queryStr += fmt.Sprintf("&offset=%d", query.Offset)
	}
	if query.Limit > 0 {
		queryStr += fmt.Sprintf("&limit=%d", query.Limit)
	}

	resp, err := http.Get("https://hypothes.is/api/search" + queryStr)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var result SearchResult
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

