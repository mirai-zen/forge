package logic

import (
	"net/http"
	"strconv"

	"github.com/mirai-zen/forge/platform/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
	"github.com/zeromicro/go-zero/rest/pathvar"

	"github.com/mirai-zen/forge-proto/platform"
)

func NewGetServiceHandler(ctx *svc.ServiceContext) *GetServiceHandler {
	return &GetServiceHandler{ctx: ctx}
}

type GetServiceHandler struct {
	ctx *svc.ServiceContext
}

func (h *GetServiceHandler) Handle(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseUint(pathvar.Vars(r)["id"], 10, 64)

	var svc struct {
		ID         uint   `gorm:"column:id"`
		Name       string `gorm:"column:name"`
		ProjectID  uint   `gorm:"column:project_id"`
		Template   string `gorm:"column:template"`
		ParamsJSON string `gorm:"column:params_json"`
		CreatedAt  string `gorm:"column:created_at"`
	}

	if err := h.ctx.DB.Table("services").Where("id = ?", id).First(&svc).Error; err != nil {
		httpx.Error(w, err)
		return
	}

	var envs []struct {
		Env       string `gorm:"column:env"`
		Namespace string `gorm:"column:namespace"`
	}
	h.ctx.DB.Table("service_envs").Where("service_id = ?", id).Scan(&envs)

	var pbEnvs []*platform.EnvStatus
	for _, e := range envs {
		pbEnvs = append(pbEnvs, &platform.EnvStatus{
			Env:       e.Env,
			Namespace: e.Namespace,
			Status:    "unknown", // TODO: 实时查 ArgoCD/K8s
		})
	}
	if pbEnvs == nil {
		pbEnvs = []*platform.EnvStatus{}
	}

	resp := &platform.GetServiceResp{
		Id:         uint64(svc.ID),
		Name:       svc.Name,
		ProjectId:  uint64(svc.ProjectID),
		Template:   svc.Template,
		ParamsJson: svc.ParamsJSON,
		CreatedAt:  svc.CreatedAt,
		Envs:       pbEnvs,
	}

	httpx.OkJson(w, resp)
}
