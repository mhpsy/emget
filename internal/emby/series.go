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

func (c *Client) GetEpisodes(ctx context.Context, seriesID, seasonID string) ([]Item, error) {
	if c.userID() == "" {
		return nil, fmt.Errorf("emby: not authenticated")
	}
	q := url.Values{}
	q.Set("SeasonId", seasonID)
	q.Set("UserId", c.userID())
	q.Set("Fields", "MediaSources,Path,Overview,ProductionYear")
	path := "/Shows/" + url.PathEscape(seriesID) + "/Episodes?" + q.Encode()

	var resp SearchResponse
	if err := c.doJSON(ctx, "GET", path, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Items, nil
}
