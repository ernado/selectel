package storage

import (
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"testing"
)

type TestClient struct {
	response *http.Response
	callback func(*http.Request) (*http.Response, error)
	err      error
}

func NewTestClientSimple(res *http.Response) doClient {
	c := new(TestClient)
	c.response = res
	return c
}

func NewTestClientError(res *http.Response, err error) doClient {
	c := new(TestClient)
	c.response = res
	c.err = err
	return c
}

func NewTestClient(callback func(*http.Request) (*http.Response, error)) doClient {
	c := new(TestClient)
	c.callback = callback
	return c
}

func (t *TestClient) Do(request *http.Request) (*http.Response, error) {
	if t.callback != nil {
		return t.callback(request)
	}
	return t.response, t.err
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
