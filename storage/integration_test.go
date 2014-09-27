package storage

import (
	"crypto/rand"
	. "github.com/smartystreets/goconvey/convey"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func randData(n int) []byte {
	const alphanum = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	var bytes = make([]byte, n)
	rand.Read(bytes)
	for i, b := range bytes {
		bytes[i] = alphanum[b%byte(len(alphanum))]
	}
	return bytes
}

func randString(n int) string {
	return string(randData(n))
}

func TestIntegration(t *testing.T) {
	test := func() {
		c, err := NewEnv()
		So(err, ShouldBeNil)
		So(c, ShouldNotBeNil)
		Convey("Async", func() {
			user := os.Getenv(EnvUser)
			key := os.Getenv(EnvKey)
			c := NewAsync(user, key)
			info := c.Info()
			So(info.BytesUsed, ShouldNotEqual, 0)
			So(info.ObjectCount, ShouldNotEqual, 0)
			So(info.ContainerCount, ShouldNotEqual, 0)
			So(func() {
				NewAsync("", key)
			}, ShouldPanic)
			So(func() {
				NewAsync("", "")
			}, ShouldPanic)
			So(func() {
				NewAsync(user, "")
			}, ShouldPanic)
			Convey("Error", func() {
				c := NewAsync(randString(10), randString(10))
				uploadData := randData(512)
				f, err := ioutil.TempFile("", randString(12))
				defer f.Close()
				f.Write(uploadData)
				So(err, ShouldBeNil)
				filename := f.Name()
				container := "test"
				So(c.UploadFile(filename, container), ShouldEqual, ErrorAuth)
			})
		})
		Convey("Info", func() {
			info := c.Info()
			So(info.BytesUsed, ShouldNotEqual, 0)
			So(info.ObjectCount, ShouldNotEqual, 0)
			So(info.ContainerCount, ShouldNotEqual, 0)
		})
		Convey("Upload", func() {
			uploadData := randData(512)
			f, err := ioutil.TempFile("", randString(12))
			defer f.Close()
			f.Write(uploadData)
			So(err, ShouldBeNil)
			filename := f.Name()
			basename := filepath.Base(filename)
			container := "test"
			So(c.UploadFile(filename, container), ShouldBeNil)
			Convey("Download", func() {
				link := c.URL(container, basename)
				req, err := http.NewRequest("GET", link, nil)
				So(err, ShouldBeNil)
				res, err := c.Do(req)
				So(err, ShouldBeNil)
				So(res.StatusCode, ShouldEqual, http.StatusOK)
				defer res.Body.Close()
				data, err := ioutil.ReadAll(res.Body)
				So(err, ShouldBeNil)
				So(string(data), ShouldEqual, string(uploadData))
				So(reflect.DeepEqual(data, uploadData), ShouldBeTrue)
				Convey("Delete", func() {
					So(c.DeleteObject(container, basename), ShouldBeNil)
					Convey("Not found", func() {
						link := c.URL(container, basename)
						req, err := http.NewRequest("GET", link, nil)
						So(err, ShouldBeNil)
						res, err := c.Do(req)
						So(err, ShouldBeNil)
						So(res.StatusCode, ShouldEqual, http.StatusNotFound)
					})
				})
			})
		})
		Convey("Container", func() {
			Convey("Upload", func() {
				uploadData := randData(512)
				f, err := ioutil.TempFile("", randString(12))
				defer f.Close()
				f.Write(uploadData)
				So(err, ShouldBeNil)
				filename := f.Name()
				basename := filepath.Base(filename)
				container := c.Container("test")
				So(container.UploadFile(filename), ShouldBeNil)
				Convey("Download", func() {
					link := c.URL(container.Name(), basename)
					req, err := http.NewRequest("GET", link, nil)
					So(err, ShouldBeNil)
					res, err := c.Do(req)
					So(err, ShouldBeNil)
					So(res.StatusCode, ShouldEqual, http.StatusOK)
					defer res.Body.Close()
					data, err := ioutil.ReadAll(res.Body)
					So(err, ShouldBeNil)
					So(string(data), ShouldEqual, string(uploadData))
					So(reflect.DeepEqual(data, uploadData), ShouldBeTrue)
					Convey("Delete", func() {
						So(container.DeleteObject(basename), ShouldBeNil)
						Convey("Not found", func() {
							link := c.URL(container.Name(), basename)
							req, err := http.NewRequest("GET", link, nil)
							So(err, ShouldBeNil)
							res, err := c.Do(req)
							So(err, ShouldBeNil)
							So(res.StatusCode, ShouldEqual, http.StatusNotFound)
						})
					})
				})
			})
		})
	}
	name := "Integration"
	if len(os.Getenv(EnvKey)) == 0 || len(os.Getenv(EnvUser)) == 0 {
		log.Println("Credentials not provided. Skipping integration tests")
		Convey(name, t, nil)
	} else {
		Convey(name, t, test)
	}
}
