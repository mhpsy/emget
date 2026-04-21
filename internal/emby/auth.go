package emby

import (
	"context"
	"time"
)

type authRequest struct {
	Username string `json:"Username"`
	Pw       string `json:"Pw"`
}

func (c *Client) Authenticate(ctx context.Context, username, password string) (*Session, error) {
	var res AuthResult
	body := authRequest{Username: username, Pw: password}
	if err := c.doJSONWithToken(ctx, "POST", "/Users/AuthenticateByName", "", body, &res); err != nil {
		return nil, err
	}
	s := &Session{
		AccessToken: res.AccessToken,
		UserID:      res.User.ID,
		ServerID:    res.ServerID,
		DeviceID:    c.deviceID,
		ExpiresAt:   time.Now().Add(30 * 24 * time.Hour), // conservative
	}
	c.SetSession(s)
	return s, nil
}
