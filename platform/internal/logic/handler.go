package logic

import (
	"net/http"

	"github.com/mirai-zen/forge/platform/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/mirai-zen/forge-proto/platform"
)

// ListTemplates — 返回可用模板列表（硬编码，后续从目录读取）

type ListTemplatesHandler struct {
	ctx *svc.ServiceContext
}

func NewListTemplatesHandler(ctx *svc.ServiceContext) *ListTemplatesHandler {
	return &ListTemplatesHandler{ctx: ctx}
}

func (h *ListTemplatesHandler) Handle(w http.ResponseWriter, r *http.Request) {
	templates := []*platform.TemplateInfo{
		{Name: "go-zero-service", Type: "service", Description: "go-zero 微服务模板", Language: "go"},
	}
	httpx.OkJson(w, &platform.ListTemplatesResp{Templates: templates})
}
