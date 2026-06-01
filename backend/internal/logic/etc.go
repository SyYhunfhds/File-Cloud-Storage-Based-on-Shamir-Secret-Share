package logic

import "github.com/gogf/gf/v2/net/ghttp"

func BeginSSE(r *ghttp.Response) {
	r.Header().Set("Content-Type", "text/event-stream")
	r.Header().Set("Cache-Control", "no-cache")
	r.Header().Set("Connection", "keep-alive")
}
