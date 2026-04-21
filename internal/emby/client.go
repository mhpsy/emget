package emby

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

var (
	ErrUnauthorized = errors.New("emby: unauthorized")
	ErrForbidden    = errors.New("emby: forbidden")
	ErrNotFound     = errors.New("emby: not found")
)

type EmbyError struct {
	Status int
	Body   string
	URL    string
}

func (e *EmbyError) Error() string {
	return fmt.Sprintf("emby: %s returned %d: %s", e.URL, e.Status, truncate(e.Body, 200))
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

type Client struct {
	base       *url.URL
	http       *http.Client
	session    *Session
	deviceID   string
	device     string
	clientName string
	version    string
}

type Session struct {
	AccessToken string
	UserID      string
	ServerID    string
	DeviceID    string
	ExpiresAt   time.Time
}

func NewClient(baseURL string, httpClient *http.Client) *Client {
	u, _ := url.Parse(baseURL)
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	host, _ := os.Hostname()
	if host == "" {
		host = "emget-host"
	}
	return &Client{
		base:       u,
		http:       httpClient,
		device:     host,
		clientName: "emget",
		version:    "0.1.0",
	}
}

func (c *Client) SetSession(s *Session) {
	c.session = s
	if s != nil {
		c.deviceID = s.DeviceID
	}
}

func (c *Client) Session() *Session { return c.session }

func (c *Client) SetDeviceID(id string) { c.deviceID = id }

func (c *Client) IsAuthenticated() bool {
	return c.session != nil && c.session.AccessToken != ""
}

func (c *Client) authHeader(token string) string {
	base := fmt.Sprintf(`MediaBrowser Client=%q, Device=%q, DeviceId=%q, Version=%q`,
		c.clientName, c.device, c.deviceID, c.version)
	if token != "" {
		return fmt.Sprintf(`MediaBrowser Token=%q, Client=%q, Device=%q, DeviceId=%q, Version=%q`,
			token, c.clientName, c.device, c.deviceID, c.version)
	}
	return base
}

func (c *Client) token() string {
	if c.session == nil {
		return ""
	}
	return c.session.AccessToken
}

func (c *Client) userID() string {
	if c.session == nil {
		return ""
	}
	return c.session.UserID
}

func (c *Client) doJSON(ctx context.Context, method, path string, body any, out any) error {
	return c.doJSONWithToken(ctx, method, path, c.token(), body, out)
}

func (c *Client) doJSONWithToken(ctx context.Context, method, path, token string, body any, out any) error {
	u, err := c.url(path)
	if err != nil {
		return err
	}
	var rdr io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		rdr = bytes.NewReader(b)
	}
	if ctx == nil {
		ctx = context.Background()
	}
	req, err := http.NewRequestWithContext(ctx, method, u, rdr)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Emby-Authorization", c.authHeader(token))
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	switch {
	case resp.StatusCode == 401:
		return ErrUnauthorized
	case resp.StatusCode == 403:
		return ErrForbidden
	case resp.StatusCode == 404:
		return ErrNotFound
	case resp.StatusCode >= 400:
		return &EmbyError{Status: resp.StatusCode, Body: string(data), URL: u}
	}

	if out != nil && len(data) > 0 {
		if err := json.Unmarshal(data, out); err != nil {
			return fmt.Errorf("emby: decode %s: %w", u, err)
		}
	}
	return nil
}

func (c *Client) url(path string) (string, error) {
	if c.base == nil {
		return "", errors.New("emby: client has no base URL")
	}
	u := *c.base
	p, q, hasQuery := strings.Cut(path, "?")
	u.Path = joinPath(u.Path, p)
	if hasQuery {
		u.RawQuery = q
	}
	return u.String(), nil
}

func joinPath(a, b string) string {
	if a == "" {
		return b
	}
	if b == "" {
		return a
	}
	if a[len(a)-1] == '/' && b[0] == '/' {
		return a + b[1:]
	}
	if a[len(a)-1] != '/' && b[0] != '/' {
		return a + "/" + b
	}
	return a + b
}
