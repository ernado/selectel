package storage

import (
	"bytes"
	. "github.com/smartystreets/goconvey/convey"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestUpload(t *testing.T) {
	c := newClient(nil)
	Convey("Upload", t, func() {
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

		Convey("Simple upload", func() {
			data := bytes.NewBufferString("data")
			callback := func(request *http.Request) (resp *http.Response, err error) {
				resp = new(http.Response)
				data, err := ioutil.ReadAll(request.Body)
				So(err, ShouldBeNil)
				So(string(data), ShouldEqual, "data")
				So(request.URL.String(), ShouldEqual, "https://xxx.selcdn.ru/container/filename")
				resp.StatusCode = http.StatusCreated
				return
			}
			c.setClient(NewTestClient(callback))

			So(c.Upload(data, "container", "filename", "text/plain"), ShouldBeNil)
		})
		Convey("404", func() {
			data := bytes.NewBufferString("data")
			callback := func(request *http.Request) (resp *http.Response, err error) {
				resp = new(http.Response)
				data, err := ioutil.ReadAll(request.Body)
				So(err, ShouldBeNil)
				So(string(data), ShouldEqual, "data")
				So(request.URL.String(), ShouldEqual, "https://xxx.selcdn.ru/container/filename")
				resp.StatusCode = http.StatusNotFound
				return
			}
			c.setClient(NewTestClient(callback))

			So(c.Upload(data, "container", "filename", "text/plain"), ShouldNotBeNil)
		})
		Convey("Url", func() {
			So(c.Container("container").URL("filename"), ShouldEqual, "https://xxx.selcdn.ru/container/filename")
			So(c.C("container").URL("filename"), ShouldEqual, "https://xxx.selcdn.ru/container/filename")
		})
	})
}
