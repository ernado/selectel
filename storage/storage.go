package storage

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"
)

const (
	authTokenHeader = "X-Auth-Token"
)

// Client is selectel storage api client
type Client struct {
	storageURL  string
	token       string
	tokenExpire int
	expireFrom  *time.Time
	user        string
	key         string
	client      doClient
}

// API for selectel storage
type API interface {
	Upload(reader io.Reader, length int64, container, filename, t string) error
	UploadFile(filename, container string) error
	Auth(user, key string) error
	Token() string
}

// doClient is mock of http.Client
type doClient interface {
	Do(request *http.Request) (*http.Response, error)
}

// setClient sets client
func (c *Client) setClient(client doClient) {
	c.client = client
}

func (c *Client) do(request *http.Request) (res *http.Response, err error) {
	if request.Header == nil {
		request.Header = http.Header{}
	}
	if request.URL.String() != authUrl && c.Expired() {
		if err = c.Auth(c.user, c.key); err != nil {
			return
		}
		if err = c.fixUrl(request); err != nil {
			return
		}
	}
	if !blank(c.token) {
		request.Header.Add(authTokenHeader, c.token)
	}
	start := time.Now().Truncate(time.Millisecond)
	res, err = c.client.Do(request)
	stop := time.Now().Truncate(time.Millisecond)
	duration := stop.Sub(start)
	if err != nil {
		log.Println(request.Method, request.URL.String(), err, duration)
		return
	}
	log.Println(request.Method, request.URL.String(), res.StatusCode, duration)
	return
}

func (c *Client) fixUrl(request *http.Request) error {
	reqUrl := request.URL
	storageUrl, urlErr := url.Parse(c.storageURL)
	if urlErr != nil {
		return urlErr
	}
	reqUrl.Host = storageUrl.Host
	reqUrl.Scheme = storageUrl.Scheme
	newRequest, err := http.NewRequest(request.Method, reqUrl.String(), request.Body)
	*request = *newRequest
	return err

}

func (c *Client) url(postfix string) string {
	return fmt.Sprintf("%s%s", c.storageURL, postfix)
}

// New returns new selectel storage api client
func New(user, key string) (API, error) {
	client := newClient(new(http.Client))
	return client, client.Auth(user, key)
}

func NewAsync(user, key string) API {
	c := newClient(new(http.Client))
	if blank(user) || blank(key) {
		panic(ErrorBadCredentials)
	}
	c.user = user
	c.key = key
	return c
}

func newClient(client *http.Client) *Client {
	c := new(Client)
	c.client = client
	return c
}

func blank(s string) bool {
	return len(s) == 0
}
