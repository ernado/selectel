package storage

import (
	"io"
)

type Container struct {
	name string
	api  API
}

type ContainerAPI interface {
	Name() string
	Upload(reader io.Reader, filename, contentType string) error
	UploadFile(filename string) error
	URL(filename string) string
}

func (c *Container) Upload(reader io.Reader, filename, contentType string) error {
	return c.api.Upload(reader, c.name, filename, contentType)
}

func (c *Container) Name() string {
	return c.name
}

func (c *Container) URL(filename string) string {
	return c.api.URL(c.name, filename)
}

func (c *Container) UploadFile(filename string) error {
	return c.api.UploadFile(filename, c.name)
}

func (c *Client) C(name string) ContainerAPI {
	container := new(Container)
	container.name = name
	container.api = c
	return container
}

func (c *Client) Container(name string) ContainerAPI {
	return c.C(name)
}
