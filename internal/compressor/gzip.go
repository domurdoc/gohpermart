package compressor

import (
	"compress/gzip"
	"io"
	"net/http"

	"github.com/domurdoc/gophermart/internal/httputil"
)

type compressWriter struct {
	w  http.ResponseWriter
	zw *gzip.Writer
	ok bool
}

func newCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{w: w, zw: gzip.NewWriter(w), ok: false}
}

func (c *compressWriter) Write(p []byte) (int, error) {
	if c.ok {
		return c.zw.Write(p)
	}
	return c.w.Write(p)
}

func (c *compressWriter) WriteHeader(statusCode int) {
	if !httputil.HasHeader(c.w.Header(), httputil.HeaderContentType) || httputil.HasContentType(
		c.w.Header(),
		httputil.ContentTypeJSON,
		httputil.ContentTypeTextPlain,
	) {
		c.ok = true
		httputil.SetContentEncoding(c.w.Header(), httputil.EncodingGZIP)
	}
	c.w.WriteHeader(statusCode)
}

func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

func (c *compressWriter) Close() error {
	return c.zw.Close()
}

type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	return &compressReader{r: r, zr: zr}, nil
}

func (c *compressReader) Read(p []byte) (int, error) {
	return c.zr.Read(p)
}

func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}

func GZIPMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ow := w
		if httputil.HasAcceptsEncoding(r.Header, httputil.EncodingGZIP) {
			cw := newCompressWriter(w)
			ow = cw
			defer cw.Close()
		}
		if httputil.HasContentEncoding(r.Header, httputil.EncodingGZIP) {
			body, err := newCompressReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			defer body.Close()
			r.Body = body
		}
		handler.ServeHTTP(ow, r)
	})
}
