package client

import (
	"encoding/json"
	"io"
	"net/url"

	"github.com/asciimoo/hister/server/indexer"
)

func (c *Client) Search(query string) (_ *indexer.Results, err error) {
	return c.SearchPage(query, "", "", false)
}

// TODO use indexer.Query as paramter
func (c *Client) SearchPage(query, pageKey, sort string, includeHTML bool) (_ *indexer.Results, err error) {
	u := "/search?q=" + url.QueryEscape(query)
	if pageKey != "" {
		u += "&page_key=" + url.QueryEscape(pageKey)
	}
	if sort != "" {
		u += "&sort=" + url.QueryEscape(sort)
	}
	if includeHTML {
		u += "&include_html=1"
	}
	req, err := c.newRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer closeBody(resp, &err)
	if err := checkStatus(resp); err != nil {
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var res *indexer.Results
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, err
	}
	return res, nil
}
