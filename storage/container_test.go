package storage

import (
	"bytes"
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"testing"
)

func TestContainerMethods(t *testing.T) {
	c := newClient(nil)
	Convey("Methods", t, func() {
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
		Convey("Name", func() {
			So(c.C("container").Name(), ShouldEqual, "container")
		})
		Convey("Url", func() {
			So(c.Container("container").URL("filename"), ShouldEqual, "https://xxx.selcdn.ru/container/filename")
			So(c.C("container").URL("filename"), ShouldEqual, "https://xxx.selcdn.ru/container/filename")
		})
		Convey("Upload", func() {
			data := bytes.NewBufferString("data")
			callback := func(request *http.Request) (resp *http.Response, err error) {
				resp = new(http.Response)
				data, err := ioutil.ReadAll(request.Body)
				So(err, ShouldBeNil)
				So(string(data), ShouldEqual, "data")
				So(request.URL.String(), ShouldEqual, "https://xxx.selcdn.ru/container/filename")
				So(request.Method, ShouldEqual, "PUT")
				resp.StatusCode = http.StatusCreated
				return
			}
			c.setClient(NewTestClient(callback))
			So(c.Container("container").Upload(data, "filename", "text/plain"), ShouldBeNil)
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
				So(request.Method, ShouldEqual, "PUT")
				So(request.URL.String(), ShouldEqual, fmt.Sprintf("https://xxx.selcdn.ru/%s/%s", container, basename))
				resp.StatusCode = http.StatusCreated
				return
			}
			c.setClient(NewTestClient(callback))
			So(c.Container(container).UploadFile(filename), ShouldBeNil)
		})
	})
}