package storage

import (
	. "github.com/smartystreets/goconvey/convey"
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
		c.setClient(&TestClient{resp, nil})
		So(c.Auth("user", "key"), ShouldBeNil)
		So(c.storageURL.String(), ShouldEqual, "https://xxx.selcdn.ru/")
		So(c.token, ShouldEqual, "token")
		So(c.tokenExpire, ShouldEqual, 110)
		Convey("Info", func() {
			resp := new(http.Response)
			resp.Header = http.Header{}
			resp.Header.Add("X-Account-Object-Count", "5563")
			resp.Header.Add("X-Account-Bytes-Used", "6427888648")
			resp.Header.Add("X-Account-Container-Count", "27")
			resp.Header.Add("X-Received-Bytes", "110278989542")
			resp.Header.Add("X-Transfered-Bytes", "224961419192")
			resp.StatusCode = http.StatusOK
			c.setClient(&TestClient{resp, nil})
			info := c.Info()
			So(info.BytesUsed, ShouldEqual, uint64(6427888648))
			So(info.ObjectCount, ShouldEqual, uint64(5563))
			So(info.ContainerCount, ShouldEqual, uint64(27))
			So(info.RecievedBytes, ShouldEqual, uint64(110278989542))
			So(info.TransferedBytes, ShouldEqual, uint64(224961419192))
		})
		Convey("Url", func() {
			So(c.url(), ShouldEqual, "https://xxx.selcdn.ru/")
			So(c.url("a"), ShouldEqual, "https://xxx.selcdn.ru/a")
			So(c.url("a", "b"), ShouldEqual, "https://xxx.selcdn.ru/a/b")
			So(c.url("a", "b", "ccc"), ShouldEqual, "https://xxx.selcdn.ru/a/b/ccc")
			So(c.URL("container", "filename"), ShouldEqual, "https://xxx.selcdn.ru/container/filename")
		})
	})
}
