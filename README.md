gentleman-log
=============

A logging plugin for [h2non/gentleman: Full-featured, plugin-oriented, composable HTTP client toolkit for Go (golang) (͡° ͜ʖ ͡°)](https://github.com/h2non/gentleman).

## Example code

```
package main

import (
	"log"
	"net/http"
	"os"

	kitlog "github.com/go-kit/kit/log"
	"github.com/h2non/gentleman"
	"github.com/h2non/gentleman/plugins/auth"
	genlog "github.com/hnakamur/gentleman-log"
	c "gopkg.in/h2non/gentleman.v1/context"
)

func main() {
	logger := kitlog.NewLogfmtLogger(os.Stdout)
	host := "http://localhost:8529"
	user := "root"
	password := "root"
	cli := gentleman.New()
	cli.Use(auth.Basic(user, password))
	logFunc := func(ctx *c.Context, req *http.Request, res *http.Response, reqBody, resBody []byte) error {
		keyvals := []interface{}{
			"method", req.Method, "url", req.URL,
		}
		for k, v := range req.Header {
			for _, vv := range v {
				keyvals = append(keyvals, "reqHeader."+k, vv)
			}
		}
		keyvals = append(keyvals,
			"payload", string(reqBody), "status", res.StatusCode)
		for k, v := range res.Header {
			for _, vv := range v {
				keyvals = append(keyvals, "resHeader."+k, vv)
			}
		}
		keyvals = append(keyvals,
			"body", string(resBody))
		logger.Log(keyvals...)
		return nil
	}
	cli.Use(genlog.Log(genlog.Config{LogFunc: logFunc}))
	cli.URL(host)
	resp, err := cli.Request().Path("/_db/_system/_api/version").BodyString(`{"dummy":"not_used"}`).Send()
	if err != nil {
		log.Printf("got error. err=%v", err)
	}
	if !resp.Ok {
		log.Printf("bad status. status=%v", resp.StatusCode)
	}
}
```

An example output.

```
$ go run main.go
method=POST url=http://localhost:8529/_db/_system/_api/version reqHeader.User-Agent=gentleman/1.0.0 reqHeader.Authorization="Basic cm9vdDpyb290" payload="{\"dummy\":\"not_used\"}" status=200 resHeader.Content-Length=37 resHeader.Server=ArangoDB resHeader.Connection=Keep-Alive resHeader.Content-Type="application/json; charset=utf-8" body="{\"server\":\"arango\",\"version\":\"3.0.3\"}"
```

## License
MIT
