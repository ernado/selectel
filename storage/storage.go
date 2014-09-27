package storage

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
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
	// EnvUser is environmental variable for selectel api username
	EnvUser = "SELECTEL_USER"
	// EnvKey is environmental variable for selectel api key
	EnvKey = "SELECTEL_KEY"
)

var (
	// ErrorObjectNotFound occurs when server returns 404
	ErrorObjectNotFound = errors.New("Object not found")
	// ErrorBadResponce occurs when server returns unexpected code
	ErrorBadResponce = errors.New("Unable to process api responce")
)

// Client is selectel storage api client
type Client struct {
	storageURL  *url.URL
	token       string
	tokenExpire int
	expireFrom  *time.Time
	user        string
	key         string
	client      DoClient
}

// StorageInformation contains some usefull metrics about storage for current user
type StorageInformation struct {
	ObjectCount     uint64
	BytesUsed       uint64
	ContainerCount  uint64
	RecievedBytes   uint64
	TransferedBytes uint64
}

// API for selectel storage
type API interface {
	DoClient
	Info() StorageInformation
	Upload(reader io.Reader, container, filename, t string) error
	UploadFile(filename, container string) error
	Auth(user, key string) error
	Token() string
	C(string) ContainerAPI
	Container(string) ContainerAPI
	DeleteObject(container, filename string) error
	URL(container, filename string) string
}

// DoClient is mock of http.Client
type DoClient interface {
	Do(request *http.Request) (*http.Response, error)
}

// setClient sets client
func (c *Client) setClient(client DoClient) {
	c.client = client
}

// DeleteObject removes object from specified container
func (c *Client) DeleteObject(container, filename string) error {
	request, err := http.NewRequest("DELETE", c.URL(container, filename), nil)
	if err != nil {
		return err
	}
	res, err := c.Do(request)
	if err != nil {
		return err
	}
	if res.StatusCode == http.StatusNotFound {
		return ErrorObjectNotFound
	}
	if res.StatusCode == http.StatusUnauthorized {
		return ErrorAuth
	}
	if res.StatusCode != http.StatusNoContent {
		return ErrorBadResponce
	}
	return nil
}

// Info returns StorageInformation for current user
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

// URL returns url for file in container
func (c *Client) URL(container, filename string) string {
	return c.url(container, filename)
}

// Do performs request with auth token
func (c *Client) Do(request *http.Request) (res *http.Response, err error) {
	return c.do(request)
}

func (c *Client) do(request *http.Request) (res *http.Response, err error) {
	if request.Header == nil {
		request.Header = http.Header{}
	}
	if request.URL.String() != authURL && c.Expired() {
		log.Println("[selectel]", "token expired")
		if err = c.Auth(c.user, c.key); err != nil {
			return
		}
		if err = c.fixURL(request); err != nil {
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

func (c *Client) fixURL(request *http.Request) error {
	newRequest, err := http.NewRequest(request.Method, c.url(request.URL.Path), request.Body)
	log.Println("fixing url", request.URL, "->", newRequest.URL.String())
	*request = *newRequest
	return err
}

func (c *Client) url(postfix ...string) string {
	path := strings.Join(postfix, "/")
	if c.storageURL == nil {
		return path
	}
	return fmt.Sprintf("%s%s", c.storageURL, path)
}

// New returns new selectel storage api client
func New(user, key string) (API, error) {
	client := newClient(new(http.Client))
	return client, client.Auth(user, key)
}

// NewAsync returns new api client and lazily performs auth
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

// NewEnv acts as New, but reads credentials from environment
func NewEnv() (API, error) {
	user := os.Getenv(EnvUser)
	key := os.Getenv(EnvKey)
	return New(user, key)
}

func blank(s string) bool {
	return len(s) == 0
}
