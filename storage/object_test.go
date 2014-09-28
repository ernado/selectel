package storage

import (
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"testing"
	"time"
)

func TestFile(t *testing.T) {
	c := newClient(nil)
	Convey("File", t, func() {
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
		Convey("Info", func() {
			Convey("Ok", func() {
				callback := func(request *http.Request) (resp *http.Response, err error) {
					resp = new(http.Response)
					So(request.URL.String(), ShouldEqual, "https://xxx.selcdn.ru/container/filename")
					So(request.Method, ShouldEqual, "HEAD")
					resp.Header = http.Header{}
					resp.Header.Set("X-Object-Downloads", "17")
					resp.Header.Set("etag", "0f343b0931126a20f133d67c2b018a3b")
					resp.Header.Set("Content-Length", "1024")
					resp.ContentLength = 1024
					resp.Header.Set("Content-Type", "application/octet-stream")
					resp.Header.Set("last-modified", "Mon, 21 May 2013 12:27:11 GMT")
					resp.StatusCode = http.StatusOK
					return
				}
				c.setClient(NewTestClient(callback))
				info, err := c.ObjectInfo("container", "filename")
				So(err, ShouldBeNil)
				So(info, ShouldNotBeNil)
				So(info.LastModified.Month(), ShouldEqual, time.May)
				So(info.LastModified.Day(), ShouldEqual, 21)
				So(info.LastModified.Year(), ShouldEqual, 2013)
				So(info.LastModified.Second(), ShouldEqual, 11)
				So(info.LastModified.Minute(), ShouldEqual, 27)
				So(info.LastModified.Hour(), ShouldEqual, 12)
				So(info.ContentType, ShouldEqual, "application/octet-stream")
				So(info.Hash, ShouldEqual, "0f343b0931126a20f133d67c2b018a3b")
				So(info.Downloaded, ShouldEqual, 17)
				So(info.Size, ShouldEqual, 1024)
				Convey("Shortcut", func() {
					info, err := c.Container("container").Object("filename").Info()
					So(err, ShouldBeNil)
					So(info, ShouldNotBeNil)
					So(info.LastModified.Month(), ShouldEqual, time.May)
					So(info.LastModified.Day(), ShouldEqual, 21)
					So(info.LastModified.Year(), ShouldEqual, 2013)
					So(info.LastModified.Second(), ShouldEqual, 11)
					So(info.LastModified.Minute(), ShouldEqual, 27)
					So(info.LastModified.Hour(), ShouldEqual, 12)
					So(info.ContentType, ShouldEqual, "application/octet-stream")
					So(info.Hash, ShouldEqual, "0f343b0931126a20f133d67c2b018a3b")
					So(info.Downloaded, ShouldEqual, 17)
					So(info.Size, ShouldEqual, 1024)
				})
			})
			Convey("Auth", func() {
				c.setClient(NewTestClientError(nil, ErrorAuth))
				_, err := c.ObjectInfo("c", "f")
				So(err, ShouldEqual, ErrorAuth)
			})
			Convey("Bad responce", func() {
				callback := func(request *http.Request) (resp *http.Response, err error) {
					resp = new(http.Response)
					So(request.URL.String(), ShouldEqual, "https://xxx.selcdn.ru/container/filename")
					So(request.Method, ShouldEqual, "HEAD")
					resp.Header = http.Header{}
					resp.Header.Set("X-Object-Downloads", "17")
					resp.Header.Set("etag", "0f343b0931126a20f133d67c2b018a3b")
					resp.Header.Set("Content-Length", "1024")
					resp.ContentLength = 1024
					resp.Header.Set("Content-Type", "application/octet-stream")
					resp.Header.Set("last-modified", "Mon, 21 May 2013 12:27:11 GMT")
					resp.StatusCode = http.StatusTeapot
					return
				}
				c.setClient(NewTestClient(callback))
				_, err := c.ObjectInfo("container", "filename")
				So(err, ShouldEqual, ErrorBadResponce)
			})
			Convey("Bad time", func() {
				callback := func(request *http.Request) (resp *http.Response, err error) {
					resp = new(http.Response)
					So(request.URL.String(), ShouldEqual, "https://xxx.selcdn.ru/container/filename")
					So(request.Method, ShouldEqual, "HEAD")
					resp.Header = http.Header{}
					resp.Header.Set("X-Object-Downloads", "17")
					resp.Header.Set("etag", "0f343b0931126a20f133d67c2b018a3b")
					resp.Header.Set("Content-Length", "1024")
					resp.ContentLength = 1024
					resp.Header.Set("Content-Type", "application/octet-stream")
					resp.Header.Set("last-modified", "asdfsafsadfsdafasdf")
					resp.StatusCode = http.StatusTeapot
					return
				}
				c.setClient(NewTestClient(callback))
				_, err := c.ObjectInfo("container", "filename")
				So(err, ShouldNotBeNil)
			})
		})
	})
}
