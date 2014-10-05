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

type fileMock interface {
	Open(name string) (*os.File, error)
	Stat(name string) (os.FileInfo, error)
}

type fileErrorMock struct {
	errOpen error
	errStat error
}

func (f fileErrorMock) Open(name string) (*os.File, error) {
	return nil, f.errOpen
}

func (f fileErrorMock) Stat(name string) (os.FileInfo, error) {
	return nil, f.errStat
}

func (c *Client) fileOpen(name string) (*os.File, error) {
	if c.file != nil {
		return c.file.Open(name)
	}
	return os.Open(name)
}

func (c *Client) fileSetMockError(errOpen, errStat error) {
	c.file = &fileErrorMock{errOpen, errStat}
}

func (c *Client) fileStat(name string) (os.FileInfo, error) {
	if c.file != nil {
		return c.file.Stat(name)
	}
	return os.Stat(name)
}

// UploadFile to container
func (c *Client) UploadFile(filename, container string) error {
	f, err := c.fileOpen(filename)
	if err != nil {
		return err
	}
	stats, err := c.fileStat(filename)
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

	request, err := c.NewRequest(putMethod, reader, container, filename)
	if err != nil {
		return err
	}
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
