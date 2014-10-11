[![Selectel](http://blog.selectel.ru/wp-content/themes/selectel/static/img/selectel.png)](https://selectel.ru/)

Selectel api
=======
```
    Language  Files   Code  Comment  Blank  Total
          Go     11   2093      132    114   2339
    Markdown      1     22        0      8     30

  Assertions: ~700
  Integrational tests included
```

[![Build Status](https://travis-ci.org/ernado/selectel.svg?branch=master)](https://travis-ci.org/ernado/selectel)
[![Coverage Status](https://img.shields.io/coveralls/ernado/selectel.svg)](https://coveralls.io/r/ernado/selectel)
[![GoDoc](https://godoc.org/github.com/ernado/selectel?status.svg)](https://godoc.org/github.com/ernado/selectel)
[![Documentation Status](https://readthedocs.org/projects/selectel-api/badge/?version=latest)](http://selectel-api.readthedocs.org/en/latest/)

Storage
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

### Console client

#### Installation

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
  --container=""  # default container (SELECTEL_CONTAINER)
  --debug         # debug mode
  -h, --help      # show help and exit
  --key=""        # selectel storage key (SELECTEL_KEY)
  --user=""       # selectel storage user (SELECTEL_USER)
  -v, --version   # show version and exit

Commands:
  info     print information about storage/container/object
```
