package log_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	kitlog "github.com/go-kit/kit/log"
	"github.com/h2non/gentleman"
	"github.com/h2non/gentleman/plugins/headers"
	genlog "github.com/hnakamur/gentleman-log"
	c "gopkg.in/h2non/gentleman.v1/context"
)

func TestLog(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)
		var req struct {
			Name string `json:"name"`
		}
		decoder.Decode(&req)

		res := struct {
			Hello string `json:"hello"`
		}{
			Hello: req.Name,
		}

		w.Header().Set("Content-Type", "application/json")
		encoder := json.NewEncoder(w)
		encoder.Encode(&res)
	}))
	defer ts.Close()

	var buf bytes.Buffer
	logger := kitlog.NewLogfmtLogger(&buf)
	host := ts.URL
	cli := gentleman.New()
	cli.Use(headers.Set("Accept", "application/json"))
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
	resp, err := cli.Request().Path("/").BodyString(`{"name":"Alice"}`).Send()
	if err != nil {
		t.Errorf("got error. err=%v", err)
	}
	if !resp.Ok {
		t.Error("resp.Ok got false; want true")
	}

	got := buf.String()
	prefix := fmt.Sprintf(`method=POST url=%s `, ts.URL)
	if !strings.HasPrefix(got, prefix) {
		t.Errorf("log %q does not have prefix %q", got, prefix)
	}
	values := []string{
		" reqHeader.User-Agent=gentleman/1.0.0 ",
		" reqHeader.Accept=application/json ",
		` payload="{\"name\":\"Alice\"}" status=200 `,
		fmt.Sprintf(` resHeader.Date="%s" `, resp.Header.Get("Date")),
		fmt.Sprintf(` resHeader.Content-Length=%s `, resp.Header.Get("Content-Length")),
		` resHeader.Content-Type=application/json `,
	}
	for _, value := range values {
		if !strings.Contains(got, value) {
			t.Errorf("log %q does not containt %q", got, value)
		}
	}
	suffix := ` body="{\"hello\":\"Alice\"}\n"` + "\n"
	if !strings.HasSuffix(got, suffix) {
		t.Errorf("log %q does not have suffix %q", got, suffix)
	}
}
