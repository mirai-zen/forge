package logic

import (
	"encoding/json"
	"net/http"

	"github.com/mirai-zen/forge-proto/platform"
	_ "github.com/mirai-zen/forge-proto/platform"
	"github.com/mirai-zen/forge/platform/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// ============================================================
// 1. CreateProject
// ============================================================

type CreateProjectHandler struct {
	ctx *svc.ServiceContext
}

func NewCreateProjectHandler(ctx *svc.ServiceContext) *CreateProjectHandler {
	return &CreateProjectHandler{ctx: ctx}
}

func (h *CreateProjectHandler) Handle(w http.ResponseWriter, r *http.Request) {
	var req platform.CreateProjectReq
	if err := httpx.Parse(r, &req); err != nil {
		httpx.Error(w, err)
		return
	}
	// TODO: call GitHub API create repo + write DB
	resp := &platform.CreateProjectResp{Message: "not implemented"}
	httpx.OkJson(w, resp)
}

// ============================================================
// 2. ListProjects
// ============================================================

type ListProjectsHandler struct {
	ctx *svc.ServiceContext
}

func NewListProjectsHandler(ctx *svc.ServiceContext) *ListProjectsHandler {
	return &ListProjectsHandler{ctx: ctx}
}

func (h *ListProjectsHandler) Handle(w http.ResponseWriter, r *http.Request) {
	var req platform.ListProjectsReq
	if err := httpx.Parse(r, &req); err != nil {
		httpx.Error(w, err)
		return
	}
	// TODO: query DB
	resp := &platform.ListProjectsResp{Projects: []*platform.ProjectInfo{}, Total: 0}
	httpx.OkJson(w, resp)
}

// ============================================================
// 3. GetProject
// ============================================================

type GetProjectHandler struct {
	ctx *svc.ServiceContext
}

func NewGetProjectHandler(ctx *svc.ServiceContext) *GetProjectHandler {
	return &GetProjectHandler{ctx: ctx}
}

func (h *GetProjectHandler) Handle(w http.ResponseWriter, r *http.Request) {
	httpx.OkJson(w, map[string]string{"message": "not implemented"})
}

// ============================================================
// 4. CreateService
// ============================================================

type CreateServiceHandler struct {
	ctx *svc.ServiceContext
}

func NewCreateServiceHandler(ctx *svc.ServiceContext) *CreateServiceHandler {
	return &CreateServiceHandler{ctx: ctx}
}

func (h *CreateServiceHandler) Handle(w http.ResponseWriter, r *http.Request) {
	var req platform.CreateServiceReq
	if err := httpx.Parse(r, &req); err != nil {
		httpx.Error(w, err)
		return
	}
	// TODO: render template + GitHub API create PR + write DB
	resp := &platform.CreateServiceResp{Message: "not implemented"}
	httpx.OkJson(w, resp)
}

// ============================================================
// 5. GetService
// ============================================================

type GetServiceHandler struct {
	ctx *svc.ServiceContext
}

func NewGetServiceHandler(ctx *svc.ServiceContext) *GetServiceHandler {
	return &GetServiceHandler{ctx: ctx}
}

func (h *GetServiceHandler) Handle(w http.ResponseWriter, r *http.Request) {
	httpx.OkJson(w, map[string]string{"message": "not implemented"})
}

// ============================================================
// 6. DeployService
// ============================================================

type DeployServiceHandler struct {
	ctx *svc.ServiceContext
}

func NewDeployServiceHandler(ctx *svc.ServiceContext) *DeployServiceHandler {
	return &DeployServiceHandler{ctx: ctx}
}

func (h *DeployServiceHandler) Handle(w http.ResponseWriter, r *http.Request) {
	var req platform.DeployServiceReq
	if err := httpx.Parse(r, &req); err != nil {
		httpx.Error(w, err)
		return
	}
	// TODO: trigger GitHub Actions workflow_dispatch
	resp := &platform.DeployServiceResp{Message: "not implemented"}
	httpx.OkJson(w, resp)
}

// ============================================================
// 7. GetEnvStatus
// ============================================================

type GetEnvStatusHandler struct {
	ctx *svc.ServiceContext
}

func NewGetEnvStatusHandler(ctx *svc.ServiceContext) *GetEnvStatusHandler {
	return &GetEnvStatusHandler{ctx: ctx}
}

func (h *GetEnvStatusHandler) Handle(w http.ResponseWriter, r *http.Request) {
	httpx.OkJson(w, map[string]string{"message": "not implemented"})
}

// ============================================================
// 8. ListTemplates
// ============================================================

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
	resp := &platform.ListTemplatesResp{Templates: templates}
	httpx.OkJson(w, resp)
}

// unused import placeholder
var _ = json.Marshal
