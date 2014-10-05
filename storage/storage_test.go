package storage

import (
	"bytes"
	. "github.com/smartystreets/goconvey/convey"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestMethods(t *testing.T) {
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
		Convey("Remove", func() {
			resp := new(http.Response)
			resp.StatusCode = http.StatusNoContent
			c.setClient(NewTestClientSimple(resp))
			So(c.RemoveObject("container", "filename"), ShouldBeNil)
			Convey("Request error", func() {
				resp := new(http.Response)
				resp.StatusCode = http.StatusNoContent
				c.setClient(NewTestClientError(resp, ErrorBadResponce))
				So(c.RemoveObject("container", "filename"), ShouldNotBeNil)
			})
			Convey("Not found", func() {
				resp := new(http.Response)
				resp.StatusCode = http.StatusNotFound
				c.setClient(NewTestClientSimple(resp))
				So(c.RemoveObject("container", "filename"), ShouldEqual, ErrorObjectNotFound)
			})
			Convey("Bad responce", func() {
				resp := new(http.Response)
				resp.StatusCode = http.StatusConflict
				c.setClient(NewTestClientError(resp, ErrorBadResponce))
				So(c.RemoveObject("container", "filename"), ShouldEqual, ErrorBadResponce)
			})
		})
		Convey("Containers", func() {
			Convey("Ok", func() {
				callback := func(req *http.Request) (*http.Response, error) {
					resp := new(http.Response)
					resp.StatusCode = http.StatusOK
					So(req.URL.String(), ShouldEqual, "https://xxx.selcdn.ru/?format=json")
					So(req.Method, ShouldEqual, "GET")
					resp.Body = ioutil.NopCloser(bytes.NewBufferString(`
					[
						{
						    "bytes": 81292837,
						    "count": 17,
						    "name": "container 1",
						    "rx_bytes": 17936289,
						    "tx_bytes": 11088956,
						    "type": "private"
						},
						{
						    "bytes": 124355568,
						    "count": 86,
						    "name": "container 2",
						    "rx_bytes": 296955064,
						    "tx_bytes": 79205328,
						    "type": "public"
						}
					]`))
					return resp, nil
				}
				c.setClient(NewTestClient(callback))
				info, err := c.Containers()
				So(err, ShouldBeNil)
				So(len(info), ShouldEqual, 2)
			})
			Convey("Auth error", func() {
				c.setClient(NewTestClientError(nil, ErrorAuth))
				_, err := c.Containers()
				So(err, ShouldEqual, ErrorAuth)
			})
		})
		Convey("ContainersInfo", func() {
			Convey("Ok", func() {
				callback := func(req *http.Request) (*http.Response, error) {
					resp := new(http.Response)
					resp.StatusCode = http.StatusOK
					So(req.URL.String(), ShouldEqual, "https://xxx.selcdn.ru/?format=json")
					So(req.Method, ShouldEqual, "GET")
					resp.Body = ioutil.NopCloser(bytes.NewBufferString(`
					[
						{
						    "bytes": 81292837,
						    "count": 17,
						    "name": "container 1",
						    "rx_bytes": 17936289,
						    "tx_bytes": 11088956,
						    "type": "private"
						},
						{
						    "bytes": 124355568,
						    "count": 86,
						    "name": "container 2",
						    "rx_bytes": 296955064,
						    "tx_bytes": 79205328,
						    "type": "public"
						}
					]`))
					return resp, nil
				}
				c.setClient(NewTestClient(callback))
				info, err := c.ContainersInfo()
				So(err, ShouldBeNil)
				So(len(info), ShouldEqual, 2)
				So(info[0].Name, ShouldEqual, "container 1")
				So(info[0].Type, ShouldEqual, "private")
				So(info[1].Name, ShouldEqual, "container 2")
				So(info[1].Type, ShouldEqual, "public")
				So(info[1].ObjectCount, ShouldEqual, 86)
			})
			Convey("Auth error", func() {
				c.setClient(NewTestClientError(nil, ErrorAuth))
				_, err := c.ContainersInfo()
				So(err, ShouldEqual, ErrorAuth)
			})
			Convey("Bad responce", func() {
				callback := func(req *http.Request) (*http.Response, error) {
					resp := new(http.Response)
					resp.StatusCode = http.StatusTeapot
					So(req.URL.String(), ShouldEqual, "https://xxx.selcdn.ru/?format=json")
					So(req.Method, ShouldEqual, "GET")
					resp.Body = ioutil.NopCloser(bytes.NewBufferString(`
					[
						{
						    "bytes": 81292837,
						    "count": 17,
						    "name": "container 1",
						    "rx_bytes": 17936289,
						    "tx_bytes": 11088956,
						    "type": "private"
						},
						{
						    "bytes": 124355568,
						    "count": 86,
						    "name": "container 2",
						    "rx_bytes": 296955064,
						    "tx_bytes": 79205328,
						    "type": "public"
						}
					]`))
					return resp, nil
				}
				c.setClient(NewTestClient(callback))
				_, err := c.ContainersInfo()
				So(err, ShouldEqual, ErrorBadResponce)
			})
			Convey("JSON error", func() {
				callback := func(req *http.Request) (*http.Response, error) {
					resp := new(http.Response)
					So(req.URL.String(), ShouldEqual, "https://xxx.selcdn.ru/?format=json")
					So(req.Method, ShouldEqual, "GET")
					resp.StatusCode = http.StatusOK
					resp.Body = ioutil.NopCloser(bytes.NewBufferString(`
					[
						{
						    "bytes": 81292837,
						    "count": 17,
						    "name": "container 1",
						    "rx_bytes": 17936289,
						    "tx_bytes": 11088956,
						    "type": "private",,,,,,,
						},
						{
						    "bytes": 124355568,
						    "count": 86,
						    "name": "container 2",
						    "rx_bytes": 296955064,
						    "tx_bytes": 79205328,
						    "type": "public"
						}
					]`))
					return resp, nil
				}
				c.setClient(NewTestClient(callback))
				_, err := c.ContainersInfo()
				So(err, ShouldNotBeNil)
			})
		})
		Convey("Info", func() {
			resp := new(http.Response)
			resp.Header = http.Header{}
			resp.Header.Add("X-Account-Object-Count", "5563")
			resp.Header.Add("X-Account-Bytes-Used", "6427888648")
			resp.Header.Add("X-Account-Container-Count", "27")
			resp.Header.Add("X-Received-Bytes", "110278989542")
			resp.Header.Add("X-Transfered-Bytes", "224961419192")
			resp.StatusCode = http.StatusOK
			c.setClient(NewTestClientSimple(resp))
			info := c.Info()
			So(info.BytesUsed, ShouldEqual, uint64(6427888648))
			So(info.ObjectCount, ShouldEqual, uint64(5563))
			So(info.ContainerCount, ShouldEqual, uint64(27))
			So(info.RecievedBytes, ShouldEqual, uint64(110278989542))
			So(info.TransferedBytes, ShouldEqual, uint64(224961419192))
			Convey("Error", func() {
				c.setClient(NewTestClientError(nil, ErrorAuth))
				info := c.Info()
				So(info.ObjectCount, ShouldEqual, uint64(0))
			})
		})
		Convey("Url", func() {
			So(c.url(), ShouldEqual, "https://xxx.selcdn.ru/")
			So(c.url("a"), ShouldEqual, "https://xxx.selcdn.ru/a")
			So(c.url("a", "b"), ShouldEqual, "https://xxx.selcdn.ru/a/b")
			So(c.url("a", "b", "ccc"), ShouldEqual, "https://xxx.selcdn.ru/a/b/ccc")
			So(c.URL("container", "filename"), ShouldEqual, "https://xxx.selcdn.ru/container/filename")
		})
		Convey("Do", func() {
			Convey("Header nil fix", func() {
				callback := func(request *http.Request) (*http.Response, error) {
					resp := new(http.Response)
					resp.Header = http.Header{}
					resp.Header.Add("X-Account-Object-Count", "5563")
					resp.Header.Add("X-Account-Bytes-Used", "6427888648")
					resp.Header.Add("X-Account-Container-Count", "27")
					resp.Header.Add("X-Received-Bytes", "110278989542")
					resp.Header.Add("X-Transfered-Bytes", "224961419192")
					resp.StatusCode = http.StatusOK
					So(request.Header, ShouldNotBeNil)
					return resp, nil
				}
				c.setClient(NewTestClient(callback))
				req, err := http.NewRequest(getMethod, "/", nil)
				req.Header = nil
				So(err, ShouldBeNil)
				res, err := c.Do(req)
				So(err, ShouldBeNil)
				So(res.StatusCode, ShouldEqual, http.StatusOK)
			})
		})
	})
}
