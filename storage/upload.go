package storage

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	// "path/filepath"
)

var (
	// ErrorUnableUpload occurs when selectel returns bad code
	ErrorUnableUpload = errors.New("Unable to upload file")
)

// UploadFile to container
func (c *Client) UploadFile(filename, container string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	stats, err := os.Stat(filename)
	if err != nil {
		return err
	}
	return c.Upload(f, stats.Size(), container, stats.Name(), "")
}

// Upload reads all data from reader and uploads to contaier with filename and content type
func (c *Client) Upload(reader io.Reader, length int64, container, filename, t string) error {
	closer, ok := reader.(io.ReadCloser)
	if ok {
		defer closer.Close()
	}

	urlStr := c.url(fmt.Sprintf("%s/%s", container, filename))
	request, err := http.NewRequest("PUT", urlStr, reader)
	if !blank(t) {
		request.Header = http.Header{}
		request.Header.Add("Content-Type", t)
	}
	if err != nil {
		return err
	}

	res, err := c.do(request)
	if err != nil {
		return err
	}

	if res.StatusCode != 201 {
		log.Printf("Bad status %s", res.Status)
		return ErrorUnableUpload
	}

	return nil
}
