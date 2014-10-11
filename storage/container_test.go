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
		Convey("Expired", func() {
			c := newClient(nil)
			So(c.Expired(), ShouldBeTrue)
		})
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
				resp := new(http.Response)
				resp.Header = http.Header{}
				resp.Header.Add("X-Container-Object-Count", "140")
				resp.Header.Add("X-Container-Bytes-Used", "68724")
				resp.Header.Add("X-Container-Meta-Type", "private")
				resp.Header.Add("X-Received-Bytes", "110278989542")
				resp.Header.Add("X-Transfered-Bytes", "224961419192")
				resp.StatusCode = http.StatusNoContent
				c.setClient(NewTestClientSimple(resp))
				info, err := c.ContainerInfo("name")
				So(err, ShouldBeNil)
				So(info.RecievedBytes, ShouldEqual, uint64(110278989542))
				So(info.TransferedBytes, ShouldEqual, uint64(224961419192))
				So(info.BytesUsed, ShouldEqual, 68724)
				So(info.Type, ShouldEqual, "private")
				So(info.ObjectCount, ShouldEqual, 140)
			})
			Convey("Ok", func() {
				/*
					Content-Length: 0
					X-Container-Object-Count: 5
					Accept-Ranges: bytes
					X-Container-Bytes-Used: 267806664
					X-Timestamp: 1386531656.46589
					X-Container-Meta-Type: public
					Content-Type: text/plain; charset=utf-8
					X-Container-Domains: sel.cydev.ru
					X-Transfered-Bytes: 424663270
					X-Received-Bytes: 477412033
					Access-Control-Allow-Origin: *
					Access-Control-Expose-Headers: X-Container-Object-Count, X-Container-Bytes-Used, X-Timestamp, X-Container-Meta-Type, X-Container-Domains, X-Transfered-Bytes, X-Received-Bytes
					Expires: 0
					Pragma: no-cache
					Cache-Control: no-cache, no-store, must-revalidate
				*/
				callback := func(req *http.Request) (*http.Response, error) {
					resp := new(http.Response)
					resp.Header = http.Header{}
					So(req.URL.String(), ShouldEqual, "https://xxx.selcdn.ru/name")
					So(req.Method, ShouldEqual, "HEAD")
					resp.Header.Add("X-Container-Object-Count", "5")
					resp.Header.Add("X-Container-Bytes-Used", "267806664")
					resp.Header.Add("X-Container-Meta-Type", "public")
					resp.Header.Add("X-Received-Bytes", "424663270")
					resp.Header.Add("X-Transfered-Bytes", "477412033")
					resp.Header.Add("Content-Length", "0")
					resp.Header.Add("X-Container-Domains", "sel.cydev.ru")
					resp.StatusCode = http.StatusNoContent
					return resp, nil
				}
				c.setClient(NewTestClient(callback))
				_, err := c.ContainerInfo("name")
				So(err, ShouldBeNil)
			})
			Convey("Bad responce", func() {
				resp := new(http.Response)
				resp.Header = http.Header{}
				resp.Header.Add("X-Container-Object-Count", "140")
				resp.Header.Add("X-Container-Bytes-Used", "68724")
				resp.Header.Add("X-Container-Meta-Type", "private")
				resp.Header.Add("X-Received-Bytes", "110278989542")
				resp.Header.Add("X-Transfered-Bytes", "224961419192")
				resp.StatusCode = http.StatusTeapot
				c.setClient(NewTestClientSimple(resp))
				_, err := c.ContainerInfo("name")
				So(err, ShouldEqual, ErrorBadResponce)
			})
			Convey("Auth", func() {
				resp := new(http.Response)
				resp.Header = http.Header{}
				resp.Header.Add("X-Container-Object-Count", "140")
				resp.Header.Add("X-Container-Bytes-Used", "68724")
				resp.Header.Add("X-Container-Meta-Type", "private")
				resp.Header.Add("X-Received-Bytes", "110278989542")
				resp.Header.Add("X-Transfered-Bytes", "224961419192")
				resp.StatusCode = http.StatusUnauthorized
				c.setClient(NewTestClientSimple(resp))
				_, err := c.ContainerInfo("name")
				So(err, ShouldEqual, ErrorAuth)
			})
			Convey("Not found", func() {
				resp := new(http.Response)
				resp.StatusCode = http.StatusNotFound
				c.setClient(NewTestClientSimple(resp))
				_, err := c.ContainerInfo("name")
				So(err, ShouldEqual, ErrorObjectNotFound)
			})
			Convey("Bad name", func() {
				resp := new(http.Response)
				resp.Header = http.Header{}
				resp.Header.Add("X-Container-Object-Count", "140")
				resp.Header.Add("X-Container-Bytes-Used", "68724")
				resp.Header.Add("X-Container-Meta-Type", "private")
				resp.Header.Add("X-Received-Bytes", "110278989542")
				resp.Header.Add("X-Transfered-Bytes", "224961419192")
				resp.StatusCode = http.StatusTeapot
				c.setClient(NewTestClientSimple(resp))
				_, err := c.ContainerInfo(randString(512))
				So(err, ShouldEqual, ErrorBadName)
			})
		})
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
		Convey("Delete object", func() {
			Convey("Ok", func() {
				callback := func(request *http.Request) (resp *http.Response, err error) {
					resp = new(http.Response)
					So(err, ShouldBeNil)
					So(request.URL.String(), ShouldEqual, "https://xxx.selcdn.ru/container/filename")
					So(request.Method, ShouldEqual, "DELETE")
					resp.StatusCode = http.StatusNoContent
					return
				}
				c.setClient(NewTestClient(callback))
				So(c.Container("container").RemoveObject("filename"), ShouldBeNil)
				So(c.Container("container").Object("filename").Remove(), ShouldBeNil)
			})
			Convey("Auth", func() {
				c.setClient(NewTestClientError(nil, ErrorAuth))
				So(c.Container("c").RemoveObject("f"), ShouldEqual, ErrorAuth)
			})
			Convey("Not found", func() {
				callback := func(request *http.Request) (resp *http.Response, err error) {
					resp = new(http.Response)
					So(err, ShouldBeNil)
					So(request.URL.String(), ShouldEqual, "https://xxx.selcdn.ru/container/filename")
					So(request.Method, ShouldEqual, "DELETE")
					resp.StatusCode = http.StatusNotFound
					return
				}
				c.setClient(NewTestClient(callback))
				So(c.Container("container").RemoveObject("filename"), ShouldEqual, ErrorObjectNotFound)
				So(c.Container("container").Object("filename").Remove(), ShouldEqual, ErrorObjectNotFound)
			})
			Convey("Bad responce", func() {
				callback := func(request *http.Request) (resp *http.Response, err error) {
					resp = new(http.Response)
					So(err, ShouldBeNil)
					So(request.URL.String(), ShouldEqual, "https://xxx.selcdn.ru/container/filename")
					So(request.Method, ShouldEqual, "DELETE")
					resp.StatusCode = http.StatusTeapot
					return
				}
				c.setClient(NewTestClient(callback))
				So(c.Container("container").RemoveObject("filename"), ShouldEqual, ErrorBadResponce)
			})
		})
		Convey("Remove", func() {
			Convey("Ok", func() {
				name := randString(10)
				callback := func(request *http.Request) (resp *http.Response, err error) {
					resp = new(http.Response)
					So(request.URL.String(), ShouldEqual, "https://xxx.selcdn.ru/"+name)
					So(request.Method, ShouldEqual, "DELETE")
					resp.StatusCode = http.StatusNoContent
					return
				}
				c.setClient(NewTestClient(callback))
				So(c.RemoveContainer(name), ShouldBeNil)
				So(c.Container(name).Remove(), ShouldBeNil)
			})
			Convey("Not empty", func() {
				name := randString(10)
				callback := func(request *http.Request) (resp *http.Response, err error) {
					resp = new(http.Response)
					So(request.URL.String(), ShouldEqual, "https://xxx.selcdn.ru/"+name)
					So(request.Method, ShouldEqual, "DELETE")
					resp.StatusCode = http.StatusConflict
					return
				}
				c.setClient(NewTestClient(callback))
				So(c.RemoveContainer(name), ShouldEqual, ErrorConianerNotEmpty)
				So(c.Container(name).Remove(), ShouldEqual, ErrorConianerNotEmpty)
			})
			Convey("Bad responce", func() {
				name := randString(10)
				callback := func(request *http.Request) (resp *http.Response, err error) {
					resp = new(http.Response)
					So(request.URL.String(), ShouldEqual, "https://xxx.selcdn.ru/"+name)
					So(request.Method, ShouldEqual, "DELETE")
					resp.StatusCode = http.StatusForbidden
					return
				}
				c.setClient(NewTestClient(callback))
				So(c.RemoveContainer(name), ShouldEqual, ErrorBadResponce)
				So(c.Container(name).Remove(), ShouldEqual, ErrorBadResponce)
			})
			Convey("Not found", func() {
				name := randString(10)
				callback := func(request *http.Request) (resp *http.Response, err error) {
					resp = new(http.Response)
					So(request.URL.String(), ShouldEqual, "https://xxx.selcdn.ru/"+name)
					So(request.Method, ShouldEqual, "DELETE")
					resp.StatusCode = http.StatusNotFound
					return
				}
				c.setClient(NewTestClient(callback))
				So(c.RemoveContainer(name), ShouldEqual, ErrorObjectNotFound)
				So(c.Container(name).Remove(), ShouldEqual, ErrorObjectNotFound)
			})
			Convey("Auth", func() {
				name := randString(10)
				callback := func(request *http.Request) (resp *http.Response, err error) {
					resp = new(http.Response)
					resp.StatusCode = http.StatusUnauthorized
					return
				}
				c.setClient(NewTestClient(callback))
				So(c.RemoveContainer(name), ShouldEqual, ErrorAuth)
				So(c.Container(name).Remove(), ShouldEqual, ErrorAuth)
			})
			Convey("Bad name", func() {
				So(c.RemoveContainer(randString(512)), ShouldEqual, ErrorBadName)
			})
		})
		Convey("Create", func() {
			Convey("Bad responce", func() {
				callback := func(request *http.Request) (resp *http.Response, err error) {
					resp = new(http.Response)
					So(request.URL.String(), ShouldEqual, "https://xxx.selcdn.ru/container")
					So(request.Method, ShouldEqual, "PUT")
					resp.StatusCode = http.StatusForbidden
					return
				}
				c.setClient(NewTestClient(callback))
				container, err := c.CreateContainer("container", false)
				So(err, ShouldEqual, ErrorBadResponce)
				So(container, ShouldBeNil)
			})
			Convey("Bad Name", func() {
				_, err := c.CreateContainer(randString(512), false)
				So(err, ShouldEqual, ErrorBadName)
			})
			Convey("Already exists", func() {
				callback := func(request *http.Request) (resp *http.Response, err error) {
					resp = new(http.Response)
					So(request.URL.String(), ShouldEqual, "https://xxx.selcdn.ru/container")
					So(request.Method, ShouldEqual, "PUT")
					resp.StatusCode = http.StatusAccepted
					return
				}
				c.setClient(NewTestClient(callback))
				container, err := c.CreateContainer("container", false)
				So(err, ShouldBeNil)
				So(container.Name(), ShouldEqual, "container")
			})
			Convey("Ok", func() {
				callback := func(request *http.Request) (resp *http.Response, err error) {
					resp = new(http.Response)
					So(request.URL.String(), ShouldEqual, "https://xxx.selcdn.ru/container")
					So(request.Method, ShouldEqual, "PUT")
					resp.StatusCode = http.StatusCreated
					return
				}
				c.setClient(NewTestClient(callback))
				container, err := c.CreateContainer("container", false)
				So(err, ShouldBeNil)
				So(container.Name(), ShouldEqual, "container")
			})
			Convey("Shortcut", func() {
				Convey("Ok", func() {
					callback := func(request *http.Request) (resp *http.Response, err error) {
						resp = new(http.Response)
						So(request.URL.String(), ShouldEqual, "https://xxx.selcdn.ru/container")
						So(request.Method, ShouldEqual, "PUT")
						resp.StatusCode = http.StatusCreated
						return
					}
					c.setClient(NewTestClient(callback))
					So(c.Container("container").Create(false), ShouldBeNil)
				})
				Convey("Auth error", func() {
					c.setClient(NewTestClientError(nil, ErrorAuth))
					So(c.Container("container").Create(false), ShouldEqual, ErrorAuth)
				})
			})
			Convey("Auth error", func() {
				c.setClient(NewTestClientError(nil, ErrorAuth))
				container, err := c.CreateContainer("container", false)
				So(err, ShouldEqual, ErrorAuth)
				So(container, ShouldBeNil)
			})
			Convey("Private", func() {
				callback := func(request *http.Request) (resp *http.Response, err error) {
					resp = new(http.Response)
					So(request.URL.String(), ShouldEqual, "https://xxx.selcdn.ru/container")
					So(request.Method, ShouldEqual, "PUT")
					resp.StatusCode = http.StatusCreated
					So(request.Header.Get(containerMetaTypeHeader), ShouldEqual, "private")
					return
				}
				c.setClient(NewTestClient(callback))
				container, err := c.CreateContainer("container", true)
				So(err, ShouldBeNil)
				So(container.Name(), ShouldEqual, "container")
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
				So(request.Method, ShouldEqual, "PUT")
				So(request.URL.String(), ShouldEqual, fmt.Sprintf("https://xxx.selcdn.ru/%s/%s", container, basename))
				resp.StatusCode = http.StatusCreated
				return
			}
			c.setClient(NewTestClient(callback))
			So(c.Container(container).UploadFile(filename), ShouldBeNil)
		})
		Convey("Objects", func() {
			Convey("Auth error", func() {
				c.setClient(NewTestClientError(nil, ErrorAuth))
				_, err := c.Container("container").Objects()
				So(err, ShouldEqual, ErrorAuth)
			})
		})
	})
}
