package emby

import (
	"net/url"
	"strconv"
)

func (c *Client) VideoDownloadURL(itemID, sourceID string) string {
	if c.base == nil {
		return ""
	}
	u := *c.base
	u.Path = joinPath(u.Path, "/Items/"+url.PathEscape(itemID)+"/Download")
	q := url.Values{}
	q.Set("api_key", c.token())
	q.Set("mediaSourceId", sourceID)
	u.RawQuery = q.Encode()
	return u.String()
}

func (c *Client) SubtitleDownloadURL(itemID, sourceID string, streamIndex int, ext string) string {
	if c.base == nil {
		return ""
	}
	u := *c.base
	u.Path = joinPath(u.Path, "/Videos/"+url.PathEscape(itemID)+"/"+url.PathEscape(sourceID)+
		"/Subtitles/"+strconv.Itoa(streamIndex)+"/Stream."+ext)
	q := url.Values{}
	q.Set("api_key", c.token())
	u.RawQuery = q.Encode()
	return u.String()
}
