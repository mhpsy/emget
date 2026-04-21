package emby

import (
	"context"
	"fmt"
	"net/url"
)

func (c *Client) GetSeasons(ctx context.Context, seriesID string) ([]Item, error) {
	if c.userID() == "" {
		return nil, fmt.Errorf("emby: not authenticated")
	}
	q := url.Values{}
	q.Set("UserId", c.userID())
	path := "/Shows/" + url.PathEscape(seriesID) + "/Seasons?" + q.Encode()

	var resp SearchResponse
	if err := c.doJSON(ctx, "GET", path, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Items, nil
}
