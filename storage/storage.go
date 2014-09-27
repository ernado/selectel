package storage

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	authTokenHeader       = "X-Auth-Token"
	objectCountHeader     = "X-Account-Object-Count"
	bytesUsedHeader       = "X-Account-Bytes-Used"
	containerCountHeader  = "X-Account-Container-Count"
	recievedBytesHeader   = "X-Received-Bytes"
	transferedBytesHeader = "X-Transfered-Bytes"
)

// Client is selectel storage api client
type Client struct {
	storageURL  *url.URL
	token       string
	tokenExpire int
	expireFrom  *time.Time
	user        string
	key         string
	client      doClient
}

type StorageInformation struct {
	ObjectCount     uint64
	BytesUsed       uint64
	ContainerCount  uint64
	RecievedBytes   uint64
	TransferedBytes uint64
}

// API for selectel storage
type API interface {
	Info() StorageInformation
	Upload(reader io.Reader, container, filename, t string) error
	UploadFile(filename, container string) error
	Auth(user, key string) error
	Token() string
	C(string) ContainerAPI
	Container(string) ContainerAPI
	URL(container, filename string) string
}

// doClient is mock of http.Client
type doClient interface {
	Do(request *http.Request) (*http.Response, error)
}

// setClient sets client
func (c *Client) setClient(client doClient) {
	c.client = client
}

func (c *Client) Info() (info StorageInformation) {
	request, err := http.NewRequest("GET", c.url(), nil)
	if err != nil {
		return
	}

	res, err := c.do(request)
	if err != nil {
		return
	}

	var (
		bitSize = 64
		base    = 10
	)

	parse := func(key string) uint64 {
		v, _ := strconv.ParseUint(res.Header.Get(key), base, bitSize)
		return v
	}

	info.BytesUsed = parse(bytesUsedHeader)
	info.ObjectCount = parse(objectCountHeader)
	info.ContainerCount = parse(containerCountHeader)
	info.RecievedBytes = parse(recievedBytesHeader)
	info.TransferedBytes = parse(transferedBytesHeader)

	return
}

func (c *Client) URL(container, filename string) string {
	return c.url(container, filename)
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
	newRequest, err := http.NewRequest(request.Method, c.url(request.URL.Path), request.Body)
	*request = *newRequest
	return err
}

func (c *Client) url(postfix ...string) string {
	return fmt.Sprintf("%s%s", c.storageURL, strings.Join(postfix, "/"))
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
