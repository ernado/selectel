package storage

import (
	"crypto/rand"
	. "github.com/smartystreets/goconvey/convey"
	"io/ioutil"
	"log"
	"net/http"
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

func TestAuth(t *testing.T) {
	Convey("Auth", t, func() {
		c, err := storage.NewEnv()
		So(err, ShouldBeNil)
		So(c, ShouldNotBeNil)
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
				log.Println("GET", link)
				res, err := http.Get(link)
				So(err, ShouldBeNil)
				So(res.StatusCode, ShouldEqual, http.StatusOK)
				defer res.Body.Close()
				data, err := ioutil.ReadAll(res.Body)
				So(err, ShouldBeNil)
				So(string(data), ShouldEqual, string(uploadData))
				So(reflect.DeepEqual(data, uploadData), ShouldBeTrue)
			})
		})
	})
}
