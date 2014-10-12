[![Selectel](http://blog.selectel.ru/wp-content/themes/selectel/static/img/selectel.png)](https://selectel.ru/)

Selectel api
=======
```
    Language  Files   Code  Comment  Blank  Total
          Go     12   2554      155    157   2866
    Markdown      1     82        0     25     98

  Assertions: ~747
  Integrational tests included
```

[![Build Status](https://travis-ci.org/ernado/selectel.svg?branch=master)](https://travis-ci.org/ernado/selectel)
[![Coverage Status](https://img.shields.io/coveralls/ernado/selectel.svg)](https://coveralls.io/r/ernado/selectel)
[![GoDoc](https://godoc.org/github.com/ernado/selectel?status.svg)](https://godoc.org/github.com/ernado/selectel)
[![Documentation Status](https://readthedocs.org/projects/selectel-api/badge/?version=latest)](http://selectel-api.readthedocs.org/en/latest/)
[![Go Walker](http://gowalker.org/api/v1/badge)](https://gowalker.org/github.com/ernado/selectel)


Selectel Storage API
=======

### Installation
```bash
go get github.com/ernado/selectel/storage
```

### Example 
```go
package main

import (
	"fmt"
	"github.com/ernado/selectel/storage"
	"log"
)

const (
	user = "123456"
	key  = "password"
)

func main() {
	api, err := storage.New(user, key)
	if err != nil {
		log.Fatal(err)
	}

	info := api.Info()
	fmt.Printf("Used %d bytes\n", info.BytesUsed)

	containers, _ := api.Containers()
	fmt.Printf("You have %d containers\n", len(containers))

	for _, container := range containers {
		objects, _ := container.Objects()
		fmt.Printf("Container %s has %d objects\n", container.Name(), len(objects))
	}
}

```

### Selectel Storage console client

#### Installation

[![Gobuild Download](http://gobuild.io/badge/github.com/ernado/selectel/storage/selctl/downloads.svg)](http://gobuild.io/github.com/ernado/selectel/storage/selctl)

```bash
go get github.com/ernado/selectel/storage/selctl	
```

#### Usage

```bash
$ selctl -h

Usage:
  selctl [options...] <command> [arg...]

Selectel storage command line client

Options:
  --cache             # cache credentials in file (SELECTEL_CACHE)
  --cache.secure      # encrypt/decrypt token with user-key pair
  -c, --container=""  # default container (SELECTEL_CONTAINER)
  --debug             # debug mode
  -h, --help          # show help and exit
  -k, --key=""        # selectel storage key (SELECTEL_KEY)
  -u, --user=""       # selectel storage user (SELECTEL_USER)
  -v, --version       # show version and exit

Commands:
  upload       upload object to container
  download     download object from container
  create       create container
  remove       remove object or container
  info         print information about storage/container/object
  list         list objects in container/storage

$ selctl info cydev main.go
# {Size:304 ContentType:application/octet-stream Downloaded:0
# Hash:f9126007fe5ac982caa9b86ad06158a9 
# LastModifiedStr: LastModified:2014-09-21 00:39:25 +0000 GMT 
# Name:main.go}

```

