package httputil

import (
	"net/http"
	"strings"
)

const (
	ContentTypeJSON          = "application/json"
	ContentTypeTextPlain     = "text/plain"
	ContentTypeTextPlainUTF8 = "text/plain; charset=utf-8"
)
const (
	EncodingGZIP = "gzip"
)
const (
	HeaderContentType     = "Content-Type"
	HeaderContentEncoding = "Content-Encoding"
	HeaderAcceptEncoding  = "Accept-Encoding"
)

func HasHeader(headers http.Header, header string) bool {
	return headers.Get(header) != ""
}

func HasContentType(headers http.Header, contentTypes ...string) bool {
	return hasHeaderValue(headers, HeaderContentType, contentTypes...)
}

func HasAcceptsEncoding(headers http.Header, encodings ...string) bool {
	return hasHeaderValue(headers, HeaderAcceptEncoding, encodings...)
}

func HasContentEncoding(headers http.Header, encodings ...string) bool {
	return hasHeaderValue(headers, HeaderContentEncoding, encodings...)
}

func hasHeaderValue(headers http.Header, headerKey string, headerValues ...string) bool {
	headerValue := strings.ToLower(headers.Get(headerKey))
	for _, v := range headerValues {
		if strings.Contains(headerValue, strings.ToLower(v)) {
			return true
		}
	}
	return false
}

func SetContentType(headers http.Header, contentType string) {
	headers.Set(HeaderContentType, contentType)
}

func SetContentEncoding(headers http.Header, encoding string) {
	headers.Set(HeaderContentEncoding, encoding)
}
