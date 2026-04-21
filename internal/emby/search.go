package emby

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func (c *Client) Search(ctx context.Context, term string, types []ItemType, limit int) ([]Item, error) {
	if c.userID() == "" {
		return nil, fmt.Errorf("emby: not authenticated")
	}
	if limit <= 0 {
		limit = 50
	}
	typeStrs := make([]string, len(types))
	for i, t := range types {
		typeStrs[i] = string(t)
	}
	q := url.Values{}
	q.Set("SearchTerm", term)
	q.Set("IncludeItemTypes", strings.Join(typeStrs, ","))
	q.Set("Recursive", "true")
	q.Set("Limit", strconv.Itoa(limit))
	q.Set("Fields", "ProductionYear,Overview")

	path := "/Users/" + url.PathEscape(c.userID()) + "/Items"
	baseURL, err := c.url(path)
	if err != nil {
		return nil, err
	}

	// Parse the base URL and add query parameters
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	u.RawQuery = q.Encode()
	fullURL := u.String()

	if ctx == nil {
		ctx = context.Background()
	}
	req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Emby-Authorization", c.authHeader(c.token()))

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	switch {
	case resp.StatusCode == 401:
		return nil, ErrUnauthorized
	case resp.StatusCode == 403:
		return nil, ErrForbidden
	case resp.StatusCode == 404:
		return nil, ErrNotFound
	case resp.StatusCode >= 400:
		return nil, &EmbyError{Status: resp.StatusCode, Body: string(data), URL: fullURL}
	}

	var searchResp SearchResponse
	if len(data) > 0 {
		if err := json.Unmarshal(data, &searchResp); err != nil {
			return nil, fmt.Errorf("emby: decode %s: %w", fullURL, err)
		}
	}
	return searchResp.Items, nil
}
