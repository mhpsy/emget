package emby

import (
	"context"
	"fmt"
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

	path := "/Users/" + url.PathEscape(c.userID()) + "/Items?" + q.Encode()

	var resp SearchResponse
	if err := c.doJSON(ctx, "GET", path, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Items, nil
}
