package storage

import (
	"errors"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
)

const (
	contentTypeHeader = "Content-Type"
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
	ext := filepath.Ext(filename)
	mimetype := mime.TypeByExtension(ext)
	return c.Upload(f, container, stats.Name(), mimetype)
}

// Upload reads all data from reader and uploads to contaier with filename and content type
func (c *Client) Upload(reader io.Reader, container, filename, contentType string) error {
	closer, ok := reader.(io.ReadCloser)
	if ok {
		defer closer.Close()
	}

	request, _ := http.NewRequest("PUT", c.URL(container, filename), reader)
	if !blank(contentType) {
		request.Header = http.Header{}
		request.Header.Add(contentTypeHeader, contentType)
	}

	res, err := c.do(request)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		return ErrorBadResponce
	}

	return nil
}
