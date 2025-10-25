package openrouter

import (
	"net/http"
)

type Option func(*Client)

func WithAuth(token string) Option {
	return func(c *Client) {
		next := c.doFunc
		c.doFunc = func(c *Client, req *http.Request) (*http.Response, error) {
			req.Header.Set("Authorization", "Bearer "+token)
			return next(c, req)
		}
	}
}
