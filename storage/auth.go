package storage

import (
	"errors"
	"net/http"
	"strconv"
)

var (
	// ErrorAuth occurs when client is unable to authenticate
	ErrorAuth = errors.New("Authentication error")
)

// Auth performs authentication to selectel and stores token and storage url
func (c *Client) Auth(user, key string) error {
	request, err := http.NewRequest("GET", "https://auth.selcdn.ru", nil)
	if err != nil {
		return err
	}
	request.Header.Add("X-Auth-User", user)
	request.Header.Add("X-Auth-Key", key)

	res, err := c.client.Do(request)
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusNoContent {
		return ErrorAuth
	}
	expire, err := strconv.Atoi(res.Header.Get("X-Expire-Auth-Token"))
	if err != nil {
		return err
	}

	c.tokenExpire = expire
	c.token = res.Header.Get("X-Auth-Token")
	if c.token == "" {
		return ErrorAuth
	}
	c.storageURL = res.Header.Get("X-Storage-Url")
	if c.storageURL == "" {
		return ErrorAuth
	}

	return nil
}
