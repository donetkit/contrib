package com_http

import (
	"bytes"
	"context"
	"crypto/tls"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"time"
)

// HTTPClient is the interface for a http client.
type HTTPClient interface {
	// Do send an HTTP request and returns an HTTP response.
	// Should use context to specify the timeout for request.
	Do(ctx context.Context, method, reqURL string, body []byte, options ...Option) (*http.Response, error)

	// Upload issues a UPLOAD to the specified URL.
	// Should use context to specify the timeout for request.
	Upload(ctx context.Context, reqURL string, form UploadForm, options ...Option) (*http.Response, error)
}

type httpclient struct {
	client *http.Client
}

func (c *httpclient) Do(ctx context.Context, method, reqURL string, body []byte, options ...Option) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, reqURL, bytes.NewBuffer(body))

	if err != nil {
		return nil, err
	}

	setting := new(config)

	if len(options) != 0 {
		setting.headers = make(map[string]string)

		for _, f := range options {
			f(setting)
		}
	}

	// headers
	if len(setting.headers) != 0 {
		for k, v := range setting.headers {
			req.Header.Set(k, v)
		}
	}

	// cookies
	if len(setting.cookies) != 0 {
		for _, v := range setting.cookies {
			req.AddCookie(v)
		}
	}

	if setting.close {
		req.Close = true
	}

	resp, err := c.client.Do(req)

	if err != nil {
		// If the context has been canceled, the context's error is probably more useful.
		select {
		case <-ctx.Done():
			err = ctx.Err()
		default:
		}

		return nil, err
	}

	return resp, err
}

func (c *httpclient) Upload(ctx context.Context, reqURL string, form UploadForm, options ...Option) (*http.Response, error) {
	buf := bytes.NewBuffer(make([]byte, 0, 20<<10)) // 20kb
	w := multipart.NewWriter(buf)

	if err := form.Write(w); err != nil {
		return nil, err
	}

	options = append(options, WithHeader("Content-Type", w.FormDataContentType()))

	// Don't forget to close the multipart writer.
	// If you don't close it, your request will be missing the terminating boundary.
	if err := w.Close(); err != nil {
		return nil, err
	}

	return c.Do(ctx, http.MethodPost, reqURL, buf.Bytes(), options...)
}

// NewHTTPClient returns a new http client
func NewHTTPClient(client *http.Client) HTTPClient {
	return &httpclient{
		client: client,
	}
}

// defaultHTTPClient default http client
var defaultHTTPClient = NewHTTPClient(&http.Client{
	Transport: &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 60 * time.Second,
		}).DialContext,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		MaxIdleConns:          0,
		MaxIdleConnsPerHost:   1000,
		MaxConnsPerHost:       1000,
		IdleConnTimeout:       60 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	},
})

// Get issues a GET to the specified URL.
func Get(ctx context.Context, reqURL string, options ...Option) (*http.Response, error) {
	return defaultHTTPClient.Do(ctx, http.MethodGet, reqURL, nil, options...)
}

// Post issues a POST to the specified URL.
func Post(ctx context.Context, reqURL string, body []byte, options ...Option) (*http.Response, error) {
	return defaultHTTPClient.Do(ctx, http.MethodPost, reqURL, body, options...)
}

// PostForm issues a POST to the specified URL, with data's keys and values URL-encoded as the request body.
func PostForm(ctx context.Context, reqURL string, data url.Values, options ...Option) (*http.Response, error) {
	options = append(options, WithHeader("Content-Type", "application/x-www-form-urlencoded"))

	return defaultHTTPClient.Do(ctx, http.MethodPost, reqURL, []byte(data.Encode()), options...)
}

// Upload issues a UPLOAD to the specified URL.
func Upload(ctx context.Context, reqURL string, form UploadForm, options ...Option) (*http.Response, error) {
	return defaultHTTPClient.Upload(ctx, reqURL, form, options...)
}

// Do send an HTTP request and returns an HTTP response
func Do(ctx context.Context, method, reqURL string, body []byte, options ...Option) (*http.Response, error) {
	return defaultHTTPClient.Do(ctx, method, reqURL, body, options...)
}
