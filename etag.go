package etag

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"hash"
	"hash/crc32"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type (
	// Config defines the config for Etag middleware.
	Config struct {
		// Skipper defines a function to skip middleware.
		Skipper middleware.Skipper
		// Weak defines if the Etag is weak or strong.
		Weak bool
		// HashFn defines the hash function to use. Default is crc32q.
		HashFn func(config Config) hash.Hash
	}
)

var (
	// DefaultEtagConfig is the default Etag middleware config.
	DefaultEtagConfig = Config{
		Skipper: middleware.DefaultSkipper,
		Weak:    true,
		HashFn: func(config Config) hash.Hash {
			if config.Weak {
				const crcPol = 0xD5828281
				crc32qTable := crc32.MakeTable(crcPol)
				return crc32.New(crc32qTable)
			}
			return sha1.New()
		},
	}
	normalizedETagName        = http.CanonicalHeaderKey("Etag")
	normalizedIfNoneMatchName = http.CanonicalHeaderKey("If-None-Match")
	weakPrefix                = "W/"
)

// Etag returns a Etag middleware.
func Etag() echo.MiddlewareFunc {
	c := DefaultEtagConfig
	return WithConfig(c)
}

// WithConfig returns a Etag middleware with config.
func WithConfig(config Config) echo.MiddlewareFunc {

	if config.Skipper == nil {
		config.Skipper = DefaultEtagConfig.Skipper
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			skipper := config.Skipper
			if skipper == nil {
				skipper = DefaultEtagConfig.Skipper
			}

			if skipper(c) {
				return next(c)
			}

			// get the hash function
			hashFn := config.HashFn
			if hashFn == nil {
				hashFn = DefaultEtagConfig.HashFn
			}

			originalWriter := c.Response().Writer
			res := c.Response()
			req := c.Request()
			// ResponseWriter
			hw := bufferedWriter{rw: res.Writer, hash: hashFn(config), buf: bytes.NewBuffer(nil)}
			res.Writer = &hw
			err = next(c)
			// restore the original writer
			res.Writer = originalWriter
			if err != nil {
				return err
			}

			resHeader := res.Header()

			if hw.hash == nil ||
				resHeader.Get(normalizedETagName) != "" ||
				strconv.Itoa(hw.status)[0] != '2' ||
				hw.status == http.StatusNoContent ||
				hw.buf.Len() == 0 {
				writeRaw(originalWriter, hw)
				return
			}

			etag := fmt.Sprintf("\"%v-%v\"", strconv.Itoa(hw.len),
				hex.EncodeToString(hw.hash.Sum(nil)))

			if config.Weak {
				etag = weakPrefix + etag
			}

			resHeader.Set(normalizedETagName, etag)

			ifNoneMatch := req.Header.Get(normalizedIfNoneMatchName) // get the If-None-Match header
			headerFresh := ifNoneMatch == etag || ifNoneMatch == weakPrefix+etag

			if headerFresh {
				originalWriter.WriteHeader(http.StatusNotModified)
				originalWriter.Write(nil)
			} else {
				writeRaw(originalWriter, hw)
			}
			return
		}
	}
}

// bufferedWriter is a wrapper around http.ResponseWriter that
// buffers the response and calculates the hash of the response.
type bufferedWriter struct {
	rw     http.ResponseWriter
	hash   hash.Hash
	buf    *bytes.Buffer
	len    int
	status int
}

// Header returns the header map that will be sent by
func (hw bufferedWriter) Header() http.Header {
	return hw.rw.Header()
}

// WriteHeader sends an HTTP response header with the provided status code.
func (hw *bufferedWriter) WriteHeader(status int) {
	hw.status = status
}

// Write writes the data to the buffer to be sent as part of an HTTP reply.
func (hw *bufferedWriter) Write(b []byte) (int, error) {
	if hw.status == 0 {
		hw.status = http.StatusOK
	}
	// write to the buffer
	l, err := hw.buf.Write(b)
	if err != nil {
		return l, err
	}
	// write to the hash
	l, err = hw.hash.Write(b)
	hw.len += l
	return l, err
}

// WriteTo writes the buffered data to the underlying io.Writer.
func writeRaw(res http.ResponseWriter, hw bufferedWriter) {
	res.WriteHeader(hw.status)
	res.Write(hw.buf.Bytes())
}
