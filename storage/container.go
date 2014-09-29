package storage

import (
	"errors"
	"io"
	"log"
	"net/http"
)

const (
	containerMetaTypeHeader = "X-Container-Meta-Type"
	containerPublic         = "public"
	containerPrivate        = "private"
)

var (
	// ErrorConianerNotEmpty occurs when requested container is not empty
	ErrorConianerNotEmpty = errors.New("Unable to remove container with objects")
)

// Container is realization of ContainerAPI
type Container struct {
	name string
	api  API
}

// ContainerAPI is interface for selectel storage container
type ContainerAPI interface {
	Name() string
	Upload(reader io.Reader, name, contentType string) error
	UploadFile(filename string) error
	URL(filename string) string
	RemoveObject(name string) error
	// Remove removes current container
	Remove() error
	// Create creates current container
	Create(bool) error
	// ObjectInfo returns info about object in container
	ObjectInfo(name string) (ObjectInfo, error)
	// Object returns object from container
	Object(name string) ObjectAPI
}

// Upload reads all data from reader and uploads to contaier with filename and content type
// shortcut to API.Upload
func (c *Container) Upload(reader io.Reader, filename, contentType string) error {
	return c.api.Upload(reader, c.name, filename, contentType)
}

// Name returns container name
func (c *Container) Name() string {
	return c.name
}

// Remove removes current container
func (c *Container) Remove() error {
	return c.api.RemoveContainer(c.name)
}

// Create creates current container
func (c *Container) Create(private bool) error {
	container, err := c.api.CreateContainer(c.name, private)
	if err != nil {
		return err
	}
	*c = *container.(*Container)
	return nil
}

// URL returns url for object
func (c *Container) URL(filename string) string {
	return c.api.URL(c.name, filename)
}

// UploadFile to current container. Shortcut to API.UploadFile
func (c *Container) UploadFile(filename string) error {
	return c.api.UploadFile(filename, c.name)
}

// DeleteObject is shortcut to API.DeleteObject
func (c *Container) RemoveObject(filename string) error {
	return c.api.RemoveObject(c.name, filename)
}

func (c *Container) ObjectInfo(name string) (ObjectInfo, error) {
	return c.api.ObjectInfo(c.name, name)
}

func (c *Container) Object(name string) ObjectAPI {
	object := new(Object)
	object.api = c.api
	object.container = c
	object.name = name
	return object
}

// C is shortcut to Client.Container
func (c *Client) C(name string) ContainerAPI {
	container := new(Container)
	container.name = name
	container.api = c
	return container
}

// Container returns new ContainerAPI client binted to container name
// Does no checks for container existance
func (c *Client) Container(name string) ContainerAPI {
	return c.C(name)
}

// CreateContainer creates new container and retuns it.
// If container already exists, function will return existing container
func (c *Client) CreateContainer(name string, private bool) (ContainerAPI, error) {
	req, err := c.NewRequest(putMethod, nil, name)
	if err != nil {
		return nil, err
	}
	req.Header = http.Header{}
	containerType := containerPublic
	if private {
		containerType = containerPrivate
	}
	req.Header.Add(containerMetaTypeHeader, containerType)
	log.Println(containerMetaTypeHeader, containerType)
	res, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode == http.StatusCreated || res.StatusCode == http.StatusAccepted {
		return c.Container(name), nil
	}
	return nil, ErrorBadResponce
}

// RemoveContainer removes container with provided name
// Container should be empty before removing and must exist
func (c *Client) RemoveContainer(name string) error {
	req, _ := http.NewRequest(deleteMethod, c.url(name), nil)
	res, err := c.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode == http.StatusConflict {
		return ErrorConianerNotEmpty
	}
	if res.StatusCode == http.StatusNotFound {
		return ErrorObjectNotFound
	}
	if res.StatusCode == http.StatusNoContent {
		return nil
	}
	return ErrorBadResponce
}
