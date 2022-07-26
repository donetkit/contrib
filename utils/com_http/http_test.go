package com_http

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHTTPOption(t *testing.T) {
	setting := &config{
		headers: make(map[string]string),
		close:   false,
	}

	options := []Option{
		WithHeader("Content-Type", "application/x-www-form-urlencoded"),
		WithClose(),
	}

	for _, f := range options {
		f(setting)
	}

	assert.Equal(t, &config{
		headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		close: true,
	}, setting)
}
