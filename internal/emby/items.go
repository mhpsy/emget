package emby

import (
	"context"
	"fmt"
	"net/url"
)

func (c *Client) GetItem(ctx context.Context, itemID string) (*Item, error) {
	if c.userID() == "" {
		return nil, fmt.Errorf("emby: not authenticated")
	}
	q := url.Values{}
	q.Set("Fields", "MediaSources,Path,Overview,ProductionYear")
	path := "/Users/" + url.PathEscape(c.userID()) +
		"/Items/" + url.PathEscape(itemID) + "?" + q.Encode()

	var item Item
	if err := c.doJSON(ctx, "GET", path, nil, &item); err != nil {
		return nil, err
	}
	return &item, nil
}
