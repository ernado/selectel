package storage

import (
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"testing"
)

type TestClient struct {
	response *http.Response
	err      error
}

func (t *TestClient) Do(request *http.Request) (*http.Response, error) {
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
			c.setClient(&TestClient{resp, nil})
			So(c.Auth("user", "key"), ShouldBeNil)
			So(c.storageURL, ShouldEqual, "https://xxx.selcdn.ru/")
			So(c.token, ShouldEqual, "token")
			So(c.tokenExpire, ShouldEqual, 110)
		})
		Convey("No token", func() {
			resp := new(http.Response)
			resp.Header = http.Header{}
			resp.Header.Add("X-Expire-Auth-Token", "110")
			resp.Header.Add("X-Storage-Url", "https://xxx.selcdn.ru/")
			resp.StatusCode = http.StatusNoContent
			c.setClient(&TestClient{resp, nil})
			So(c.Auth("user", "key"), ShouldNotBeNil)
		})
		Convey("No url", func() {
			resp := new(http.Response)
			resp.Header = http.Header{}
			resp.Header.Add("X-Auth-Token", "token")
			resp.Header.Add("X-Expire-Auth-Token", "110")
			resp.StatusCode = http.StatusNoContent
			c.setClient(&TestClient{resp, nil})
			So(c.Auth("user", "key"), ShouldNotBeNil)
		})
		Convey("No expire", func() {
			resp := new(http.Response)
			resp.Header = http.Header{}
			resp.StatusCode = http.StatusNoContent
			c.setClient(&TestClient{resp, nil})
			So(c.Auth("user", "key"), ShouldNotBeNil)
		})
		Convey("Error is not nil", func() {
			resp := new(http.Response)
			resp.Header = http.Header{}
			resp.Header.Add("X-Expire-Auth-Token", "110")
			resp.Header.Add("X-Storage-Url", "https://xxx.selcdn.ru/")
			resp.StatusCode = http.StatusNoContent
			c.setClient(&TestClient{resp, http.ErrBodyNotAllowed})
			So(c.Auth("user", "key"), ShouldNotBeNil)
		})
		Convey("Bad code", func() {
			resp := new(http.Response)
			resp.Header = http.Header{}
			resp.Header.Add("X-Expire-Auth-Token", "110")
			resp.Header.Add("X-Storage-Url", "https://xxx.selcdn.ru/")
			resp.StatusCode = http.StatusForbidden
			c.setClient(&TestClient{resp, nil})
			So(c.Auth("user", "key"), ShouldNotBeNil)
		})
	})
}
