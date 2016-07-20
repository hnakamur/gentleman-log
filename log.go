// Package log provides a logging plugin for gopkg.in/h2non/gentleman.v1.
package log

import (
	"bytes"
	"io/ioutil"
	"net/http"

	c "gopkg.in/h2non/gentleman.v1/context"
	p "gopkg.in/h2non/gentleman.v1/plugin"
)

// Config is a config for the log plugin.
type Config struct {
	// ReqBodyKey is the key for the request body in the ctx.
	// If this is an empty, the default value "req.body" is used.
	ReqBodyKey string

	// LogFunc is the logging function.
	LogFunc func(ctx *c.Context, req *http.Request, res *http.Response, reqBody, resBody []byte) error
}

// Log logs requests and responses.
func Log(config Config) p.Plugin {
	var reqBodyKey string
	if config.ReqBodyKey == "" {
		reqBodyKey = "req.body"
	} else {
		reqBodyKey = config.ReqBodyKey
	}

	handleBeforeDial := func(ctx *c.Context, h c.Handler) {
		var body []byte
		if ctx.Request.Body != nil {
			var err error
			body, err = ioutil.ReadAll(ctx.Request.Body)
			if err != nil {
				h.Error(ctx, err)
				return
			}
			ctx.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body))
		}
		ctx.Set(reqBodyKey, body)

		h.Next(ctx)
	}

	handleResponse := func(ctx *c.Context, h c.Handler) {
		var reqBody []byte
		reqBodyVal := ctx.Get(reqBodyKey)
		if reqBodyVal != nil {
			reqBody = reqBodyVal.([]byte)
		}

		var resBody []byte
		if ctx.Response.Body != nil {
			var err error
			resBody, err = ioutil.ReadAll(ctx.Response.Body)
			if err != nil {
				h.Error(ctx, err)
				return
			}
			ctx.Response.Body = ioutil.NopCloser(bytes.NewBuffer(resBody))
		}

		err := config.LogFunc(ctx, ctx.Request, ctx.Response, reqBody, resBody)
		if err != nil {
			h.Error(ctx, err)
		}

		h.Next(ctx)
	}
	handlers := p.Handlers{
		"before dial": handleBeforeDial,
		"response":    handleResponse,
	}
	return &p.Layer{Handlers: handlers}
}
