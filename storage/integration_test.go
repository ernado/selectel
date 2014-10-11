package storage

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
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
		Convey("ContainerInfo", func() {
			_, err := c.Container("cydev").Info()
			So(err, ShouldBeNil)
		})
		Convey("Upload", func() {
			uploadData := randData(512)
			hashByte := md5.Sum(uploadData)
			hash := hex.EncodeToString(hashByte[:])
			f, err := ioutil.TempFile("", randString(12))
			defer f.Close()
			f.Write(uploadData)
			So(err, ShouldBeNil)
			filename := f.Name()
			basename := filepath.Base(filename)
			container := "test"
			So(c.UploadFile(filename, container), ShouldBeNil)
			Convey("File info", func() {
				info, err := c.ObjectInfo(container, basename)
				So(err, ShouldBeNil)
				So(info.Hash, ShouldEqual, hash)
				So(info.Downloaded, ShouldEqual, 0)
			})
			Convey("Download shortcut", func() {
				reader, err := c.Container(container).Object(basename).GetReader()
				So(err, ShouldBeNil)
				data, err := ioutil.ReadAll(reader)
				So(err, ShouldBeNil)
				So(string(data), ShouldEqual, string(uploadData))
				So(reflect.DeepEqual(data, uploadData), ShouldBeTrue)
				Convey("Remove", func() {
					So(c.Container(container).Object(basename).Remove(), ShouldBeNil)
					Convey("Not found", func() {
						_, err := c.Container(container).Object(basename).GetReader()
						So(err, ShouldEqual, ErrorObjectNotFound)
					})
				})
			})
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
				Convey("Remove", func() {
					So(c.RemoveObject(container, basename), ShouldBeNil)
					Convey("Not found", func() {
						link := c.URL(container, basename)
						req, err := http.NewRequest("GET", link, nil)
						So(err, ShouldBeNil)
						res, err := c.Do(req)
						So(err, ShouldBeNil)
						So(res.StatusCode, ShouldEqual, http.StatusNotFound)
					})
					Convey("File info", func() {
						_, err := c.ObjectInfo(container, basename)
						So(err, ShouldEqual, ErrorObjectNotFound)
					})
				})
			})
		})
		Convey("Container", func() {
			Convey("Create", func() {
				name := fmt.Sprintf("test_%s", randString(30))
				So(c.Container(name).Create(false), ShouldBeNil)
				Convey("Already exists", func() {
					So(c.Container(name).Create(false), ShouldBeNil)
				})
				Convey("Upload", func() {
					uploadData := randData(512)
					f, err := ioutil.TempFile("", randString(12))
					defer f.Close()
					f.Write(uploadData)
					So(err, ShouldBeNil)
					filename := f.Name()
					basename := filepath.Base(filename)
					container := c.Container(name)
					So(container.UploadFile(filename), ShouldBeNil)
					Convey("Remove container", func() {
						So(c.Container(name).Remove(), ShouldEqual, ErrorConianerNotEmpty)
					})
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
						Convey("Remove", func() {
							So(container.RemoveObject(basename), ShouldBeNil)
							Convey("Remove container", func() {
								So(c.Container(name).Remove(), ShouldBeNil)
							})
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
				Convey("Remove", func() {
					So(c.Container(name).Remove(), ShouldBeNil)
				})
				Convey("List", func() {
					containers, err := c.Containers()
					So(err, ShouldBeNil)
					So(len(containers), ShouldNotEqual, 0)
					for _, container := range containers {
						if !strings.Contains(container.Name(), "test_") {
							continue
						}
						log.Println("removing container")
						objects, err := container.Objects()
						So(err, ShouldBeNil)
						for _, object := range objects {
							So(object.Remove(), ShouldBeNil)
						}
						So(container.Remove(), ShouldBeNil)
					}
				})
			})
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
					Convey("Remove", func() {
						So(container.RemoveObject(basename), ShouldBeNil)
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
