package storage

import (
	"errors"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
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

	request, err := http.NewRequest("PUT", c.URL(container, filename), reader)
	if !blank(contentType) {
		request.Header = http.Header{}
		request.Header.Add("Content-Type", contentType)
	}
	if err != nil {
		return err
	}

	res, err := c.do(request)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		log.Printf("Bad status %s", res.Status)
		return ErrorUnableUpload
	}

	return nil
}
