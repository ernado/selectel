package storage

import (
	"bytes"
	. "github.com/smartystreets/goconvey/convey"
	"io/ioutil"
	"net/http"
	"testing"
)

const (
	invalidHost = "top_kek%%%%%333@eeкек."
)

type TestClient struct {
	callback func(*http.Request) (*http.Response, error)
}

func (c *TestClient) simpleCallback(res *http.Response, err error) func(*http.Request) (*http.Response, error) {
	return func(_ *http.Request) (*http.Response, error) {
		return res, err
	}
}

func NewTestClientSimple(res *http.Response) DoClient {
	c := new(TestClient)
	c.callback = c.simpleCallback(res, nil)
	return c
}

func NewTestClientError(res *http.Response, err error) DoClient {
	c := new(TestClient)
	c.callback = c.simpleCallback(res, err)
	return c
}

func NewTestClient(callback func(*http.Request) (*http.Response, error)) DoClient {
	return &TestClient{callback: callback}
}

func (c *TestClient) Do(request *http.Request) (res *http.Response, err error) {
	res, err = c.callback(request)
	if err == nil && res.Body == nil {
		res.Body = ioutil.NopCloser(bytes.NewBuffer(nil))
	}
	return res, err
}

func forceBadURL(c *Client) {
	c.storageURL.Host = invalidHost
}

func TestAuth(t *testing.T) {
	c := newClient(nil)
	Convey("Auth", t, func() {
		Convey("Ok", func() {
			resp := new(http.Response)
			resp.Header = http.Header{}
			resp.Header.Add("X-Expire-Auth-Token", "110")
			resp.Header.Add("X-Auth-Token", "token")
			resp.Header.Add("X-Storage-Url", "https://xxx.selcdn.ru/")
			resp.StatusCode = http.StatusNoContent
			c.setClient(NewTestClientSimple(resp))
			So(c.Auth("user", "key"), ShouldBeNil)
			So(c.storageURL.String(), ShouldEqual, "https://xxx.selcdn.ru/")
			So(c.token, ShouldEqual, "token")
			So(c.tokenExpire, ShouldEqual, 110)
			So(c.Token(), ShouldEqual, "token")
		})
		Convey("Bad url", func() {
			Convey("Request", func() {
				resp := new(http.Response)
				resp.Header = http.Header{}
				resp.Header.Add("X-Expire-Auth-Token", "110")
				resp.Header.Add("X-Auth-Token", "token")
				resp.Header.Add("X-Storage-Url", "https://xxx.selcdn.ru/")
				resp.StatusCode = http.StatusNoContent
				c.setClient(NewTestClientSimple(resp))
				So(c.Auth("user", "key"), ShouldBeNil)
				So(c.storageURL.String(), ShouldEqual, "https://xxx.selcdn.ru/")
				forceBadURL(c)
				_, err := c.NewRequest("GET", nil)
				So(err, ShouldNotBeNil)
			})
			Convey("Header", func() {
				resp := new(http.Response)
				resp.Header = http.Header{}
				resp.Header.Add("X-Expire-Auth-Token", "110")
				resp.Header.Add("X-Auth-Token", "token")
				resp.Header.Add("X-Storage-Url", invalidHost)
				resp.StatusCode = http.StatusNoContent
				c.setClient(NewTestClientSimple(resp))
				So(c.Auth("user", "key"), ShouldNotBeNil)
			})
		})
		Convey("Bad credentianls", func() {
			So(c.Auth("", "key"), ShouldEqual, ErrorBadCredentials)
			So(c.Auth("user", ""), ShouldEqual, ErrorBadCredentials)
			So(c.Auth("", ""), ShouldEqual, ErrorBadCredentials)
		})
		Convey("No token", func() {
			resp := new(http.Response)
			resp.Header = http.Header{}
			resp.Header.Add("X-Expire-Auth-Token", "110")
			resp.Header.Add("X-Storage-Url", "https://xxx.selcdn.ru/")
			resp.StatusCode = http.StatusNoContent
			c.setClient(NewTestClientSimple(resp))
			So(c.Auth("user", "key"), ShouldNotBeNil)
		})
		Convey("No url", func() {
			resp := new(http.Response)
			resp.Header = http.Header{}
			resp.Header.Add("X-Auth-Token", "token")
			resp.Header.Add("X-Expire-Auth-Token", "110")
			resp.StatusCode = http.StatusNoContent
			c.setClient(NewTestClientSimple(resp))
			So(c.Auth("user", "key"), ShouldNotBeNil)
		})
		Convey("No expire", func() {
			resp := new(http.Response)
			resp.Header = http.Header{}
			resp.StatusCode = http.StatusNoContent
			c.setClient(NewTestClientSimple(resp))
			So(c.Auth("user", "key"), ShouldNotBeNil)
		})
		Convey("Error is not nil", func() {
			resp := new(http.Response)
			resp.Header = http.Header{}
			resp.Header.Add("X-Expire-Auth-Token", "110")
			resp.Header.Add("X-Storage-Url", "https://xxx.selcdn.ru/")
			resp.StatusCode = http.StatusNoContent
			c.setClient(NewTestClientError(resp, http.ErrBodyNotAllowed))
			So(c.Auth("user", "key"), ShouldNotBeNil)
		})
		Convey("Bad code", func() {
			resp := new(http.Response)
			resp.Header = http.Header{}
			resp.Header.Add("X-Expire-Auth-Token", "110")
			resp.Header.Add("X-Storage-Url", "https://xxx.selcdn.ru/")
			resp.StatusCode = http.StatusForbidden
			c.setClient(NewTestClientSimple(resp))
			So(c.Auth("user", "key"), ShouldNotBeNil)
		})
	})
}
