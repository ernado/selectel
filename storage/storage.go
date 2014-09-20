package storage

import (
	"net/http"
)

// Client is selectel storage api client
type Client struct {
	storageURL  string
	token       string
	tokenExpire int
	client      DoClient
}

// DoClient is mock of http.Client
type DoClient interface {
	Do(request *http.Request) (*http.Response, error)
}

// SetClient sets client
func (c *Client) SetClient(client DoClient) {
	c.client = client
}

// New returns new selectel storage api client
func New() *Client {
	return newClient(new(http.Client))
}

func newClient(client *http.Client) *Client {
	c := new(Client)
	c.client = client
	return c
}
