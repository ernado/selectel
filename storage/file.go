package storage

import (
	"time"
)

type File struct {
	Size         uint64    `json:"bytes"`
	ContentType  string    `json:"content_type"`
	Downloaded   int       `json:"downloaded"`
	Hash         string    `json:"hash"`
	LastModified time.Time `json:"last_modified"`
	Name         string    `json:"name"`
}
