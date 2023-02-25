# Echo Etag Middleware

<!-- [![Build Status](https://travis-ci.org/foolin/echo-etag.svg?branch=master)](https://travis-ci.org/foolin/echo-etag)
[![Go Report Card](https://goreportcard.com/badge/github.com/pablor21/echo-etag)](https://goreportcard.com/report/github.com/pablor21/echo-etag)
[![GoDoc](https://godoc.org/github.com/pablor21/echo-etag?status.svg)](https://godoc.org/github.com/pablor21/echo-etag)
[![codecov](https://codecov.io/gh/foolin/echo-etag/branch/master/graph/badge.svg)](https://codecov.io/gh/foolin/echo-etag)
[![GitHub release](https://img.shields.io/github/release/foolin/echo-etag.svg)](https://img.shields.io/github/release/foolin/echo-etag.svg)
[![GitHub license](https://img.shields.io/badge/license-MIT-blue.svg)](https://raw.githubusercontent.com/foolin/echo-etag/master/LICENSE) -->


Etag middleware for Echo Framework

## Features

- Support Etag and Weak Etag
- Support Skipper
- Configurable Hash Function (default: sha256 for Etag and crc32 for Weak Etag) Any hash function that implements the hash.Hash interface can be used.



> BEWARE: Creating an Etag will buffer the entire response body. This may consume a lot of memory if the response body is large (files). If this is a concern, you should use the Skipper option to skip Etag generation for large responses and use other method of caching, for example: [Last-Modified header](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Last-Modified).



## Installation

```bash
$ go get github.com/pablor21/echo-etag
```

## Usage

```go
package main

import (
    "github.com/labstack/echo/v4"
    "github.com/pablor21/echo-etag/v4"
)

func main() {
    e := echo.New()

    //Etag middleware
    e.Use(etag.Etag())

    e.Start(":1323")
}
```

## Example

```go

package main

import (
    "github.com/labstack/echo/v4"
    "github.com/pablor21/echo-etag/v4"
)

func main() {
    e := echo.New()

    //Etag middleware
    e.Use(etag.Etag())

    e.GET("/", func(c echo.Context) error {
        return c.String(200, "Hello, World!")
    })

    e.Start(":1323")
}

```

## Configuration

```go

package main

import (
    "crypto/md5"
    "github.com/labstack/echo/v4"
    "github.com/pablor21/echo-etag/v4"
)

func main() {
    e := echo.New()

    //Etag middleware
    e.Use(etag.EtagWithConfig(etag.Config{
        Skipper: func(c echo.Context) bool {
            return c.Path() == "/skip"
        },
        Weak: true,
        HashFn: func(config EtagConfig) hash.Hash {
            return md5.New() //use md5 hash
		},
    }))

    e.GET("/", func(c echo.Context) error {
        return c.String(200, "Hello, World!")
    })

    e.Start(":1323")
}

```
