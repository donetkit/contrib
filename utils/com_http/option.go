package com_http

import "net/http"

// config http request setting
type config struct {
	headers map[string]string
	cookies []*http.Cookie
	close   bool
}

// Option configures how we set up the http request.
type Option func(s *config)

// WithHeader specifies the header to http request.
func WithHeader(key, value string) Option {
	return func(s *config) {
		s.headers[key] = value
	}
}

// WithCookies specifies the cookies to http request.
func WithCookies(cookies ...*http.Cookie) Option {
	return func(s *config) {
		s.cookies = cookies
	}
}

// WithClose specifies close the connection after
// replying to this request (for servers) or after sending this
// request and reading its response (for clients).
func WithClose() Option {
	return func(s *config) {
		s.close = true
	}
}
