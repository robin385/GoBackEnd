package handlers

import (
	"encoding/json"

	"github.com/valyala/fasthttp"
)

func writeJSON(ctx *fasthttp.RequestCtx, status int, v interface{}) {
	ctx.Response.Header.Set("Content-Type", "application/json")
	ctx.SetStatusCode(status)
	if v != nil {
		if data, err := json.Marshal(v); err == nil {
			_, _ = ctx.Write(data)
		} else {
			ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
		}
	}
}

func readJSON(ctx *fasthttp.RequestCtx, v interface{}) error {
	return json.Unmarshal(ctx.PostBody(), v)
}
