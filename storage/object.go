package storage

import (
	"net/http"
	"strconv"
	"time"
)

const (
	etagHeader            = "etag"
	contentLengthHeader   = "Content-Length"
	lastModifiedLayout    = time.RFC1123
	lastModifiedHeader    = "last-modified"
	objectDownloadsHeader = "X-Object-Downloads"
)

// ObjectInfo represents object info
type ObjectInfo struct {
	Size         uint64    `json:"bytes"`
	ContentType  string    `json:"content_type"`
	Downloaded   uint64    `json:"downloaded"`
	Hash         string    `json:"hash"`
	LastModified time.Time `json:"last_modified"`
	Name         string    `json:"name"`
}

func (c *Client) ObjectInfo(container, filename string) (f ObjectInfo, err error) {
	request, _ := http.NewRequest(headMethod, c.URL(container, filename), nil)
	res, err := c.do(request)
	if err != nil {
		return f, err
	}
	if res.StatusCode == http.StatusNotFound {
		return f, ErrorObjectNotFound
	}
	if res.StatusCode != http.StatusOK {
		return f, ErrorBadResponce
	}
	parse := func(key string) uint64 {
		v, _ := strconv.ParseUint(res.Header.Get(key), uint64Base, uint64BitSize)
		return v
	}
	f.Size = uint64(res.ContentLength)
	f.Hash = res.Header.Get(etagHeader)
	f.ContentType = res.Header.Get(contentTypeHeader)
	f.LastModified, err = time.Parse(lastModifiedLayout, res.Header.Get(lastModifiedHeader))
	f.Name = filename
	if err != nil {
		return
	}
	f.Downloaded = parse(objectDownloadsHeader)
	return
}
