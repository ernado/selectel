package storage

import (
	"bytes"
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"io"
	"io/ioutil"
	"net/http"
	"path/filepath"
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
			Convey("Ok", func() {
				callback := func(request *http.Request) (resp *http.Response, err error) {
					resp = new(http.Response)
					data, err := ioutil.ReadAll(request.Body)
					So(err, ShouldBeNil)
					So(string(data), ShouldEqual, "data")
					So(request.Header.Get("etag"), ShouldEqual, "8d777f385d3dfec8815d20f7496026dc")
					So(request.URL.String(), ShouldEqual, "https://xxx.selcdn.ru/container/filename")
					resp.StatusCode = http.StatusCreated
					return
				}
				c.setClient(NewTestClient(callback))
				Convey("Direct", func() {
					So(c.Upload(data, "container", "filename", "text/plain"), ShouldBeNil)
				})
				Convey("Shortcut", func() {
					So(c.Container("container").Object("filename").Upload(data, "text/plain"), ShouldBeNil)
				})
			})
			Convey("Not found", func() {
				callback := func(request *http.Request) (resp *http.Response, err error) {
					resp = new(http.Response)
					data, err := ioutil.ReadAll(request.Body)
					So(err, ShouldBeNil)
					So(string(data), ShouldEqual, "data")
					So(request.URL.String(), ShouldEqual, "https://xxx.selcdn.ru/container/filename")
					So(request.Method, ShouldEqual, "PUT")
					resp.StatusCode = http.StatusNotFound
					return
				}
				c.setClient(NewTestClient(callback))
				So(c.Upload(data, "container", "filename", "text/plain"), ShouldNotBeNil)
			})
			Convey("Auth", func() {
				c.setClient(NewTestClientError(nil, ErrorAuth))
				So(c.Upload(data, "c", "f", "text/plain"), ShouldEqual, ErrorAuth)
			})
			Convey("Bad name", func() {
				So(c.Upload(data, "c", randString(512), "text/plain"), ShouldNotBeNil)
				So(c.Upload(data, randString(512), randString(512), "text/plain"), ShouldNotBeNil)
				So(c.Upload(data, randString(512), "f", "text/plain"), ShouldNotBeNil)
				So(c.Upload(data, randString(512), "f", randString(1024)), ShouldNotBeNil)
			})
			Convey("Bad url", func() {
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
				forceBadURL(c)
				So(c.Upload(data, "c", "f", "text/plain"), ShouldEqual, ErrorBadName)
			})
		})
		Convey("File upload", func() {
			f, err := ioutil.TempFile("", "data")
			defer f.Close()
			f.WriteString("data")
			So(err, ShouldBeNil)
			filename := f.Name()
			basename := filepath.Base(filename)
			container := "container"
			callback := func(request *http.Request) (resp *http.Response, err error) {
				resp = new(http.Response)
				data, err := ioutil.ReadAll(request.Body)
				So(err, ShouldBeNil)
				So(string(data), ShouldEqual, "data")
				So(request.URL.String(), ShouldEqual, fmt.Sprintf("https://xxx.selcdn.ru/%s/%s", container, basename))
				So(request.Method, ShouldEqual, "PUT")
				resp.StatusCode = http.StatusCreated
				return
			}
			c.setClient(NewTestClient(callback))
			Convey("Direct", func() {
				So(c.UploadFile(filename, container), ShouldBeNil)
			})
			Convey("Shortcut", func() {
				So(c.Container("container").Object(basename).UploadFile(filename), ShouldBeNil)
			})
			Convey("IO error", func() {
				Convey("Bad file", func() {
					So(c.UploadFile(randString(100), "container"), ShouldNotBeNil)
				})
				Convey("Open error", func() {
					c.fileSetMockError(io.ErrUnexpectedEOF, nil)
					So(c.UploadFile(randString(100), "container"), ShouldNotBeNil)
				})
				Convey("Stat error", func() {
					c.fileSetMockError(nil, io.ErrUnexpectedEOF)
					So(c.UploadFile(randString(100), "container"), ShouldNotBeNil)
				})
			})
		})

		Convey("Url", func() {
			So(c.Container("container").URL("filename"), ShouldEqual, "https://xxx.selcdn.ru/container/filename")
			So(c.C("container").URL("filename"), ShouldEqual, "https://xxx.selcdn.ru/container/filename")
		})
	})
}
