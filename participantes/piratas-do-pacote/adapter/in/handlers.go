package in

import (
	"encoding/json"

	"github.com/piratas-do-pacote/core"
	"github.com/valyala/fasthttp"
)

func (h *HttpHandler) handleFindService(ctx *fasthttp.RequestCtx) {

	ctx.Response.Header.SetContentType("application/json")
	ctx.SetStatusCode(fasthttp.StatusOK)
	body := ctx.PostBody()

	var inferenceInput core.InferenceInput
	_ = json.Unmarshal(body, &inferenceInput)

	result := h.inferenceUseCase.Infer(ctx, inferenceInput)
	jsonResult, _ := json.Marshal(result)

	ctx.Response.SetBodyString(string(jsonResult))
	return
}

func (h *HttpHandler) handleHeathZ(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.SetContentType("application/json")
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.Response.SetBodyString(`{"status":"ok"}`)
	return
}
