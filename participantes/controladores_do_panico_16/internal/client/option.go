package client

import (
"fmt"
"net/http"
)

type Option func(*Client)

func WithAuth(token string) Option {
fmt.Printf("[WithAuth] Token being injected:\n")
fmt.Printf("Token: %s\n", token)
fmt.Printf("Length: %d characters\n\n", len(token))

return func(c *Client) {
next := c.doFunc
c.doFunc = func(c *Client, req *http.Request) (*http.Response, error) {
authHeader := "Bearer " + token
fmt.Printf("[Request] Adding Authorization header:\n")
fmt.Printf("Authorization: %s\n", authHeader)
fmt.Printf("Total length: %d characters\n\n", len(authHeader))
req.Header.Set("Authorization", authHeader)
return next(c, req)
}
}
}

func min(a, b int) int {
if a < b {
return a
}
return b
}
