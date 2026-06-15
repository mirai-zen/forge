package logic

import (
	"net/http"
	"strconv"

	"github.com/mirai-zen/forge/platform/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
	"github.com/zeromicro/go-zero/rest/pathvar"

	"github.com/mirai-zen/forge-proto/platform"
)

func NewGetEnvStatusHandler(ctx *svc.ServiceContext) *GetEnvStatusHandler {
	return &GetEnvStatusHandler{ctx: ctx}
}

type GetEnvStatusHandler struct {
	ctx *svc.ServiceContext
}

func (h *GetEnvStatusHandler) Handle(w http.ResponseWriter, r *http.Request) {
	serviceID, _ := strconv.ParseUint(pathvar.Vars(r)["id"], 10, 64)
	env := pathvar.Vars(r)["env"]

	// 查询 namespace
	var ns string
	h.ctx.DB.Table("service_envs").
		Where("service_id = ? AND env = ?", serviceID, env).
		Select("namespace").Scan(&ns)

	if ns == "" {
		httpx.Error(w, errNotFound(env))
		return
	}

	// TODO: 实时查询 ArgoCD/K8s API 获取部署状态
	// 当前返回占位状态
	resp := &platform.GetEnvStatusResp{
		Env:        env,
		Namespace:  ns,
		Status:     "unknown", // TODO: ArgoCD health status
		SyncStatus: "unknown", // TODO: ArgoCD sync status
		Replicas:   "0",
		Image:      "-",
	}

	httpx.OkJson(w, resp)
}

func errNotFound(env string) error {
	return &envNotFoundError{env: env}
}

type envNotFoundError struct {
	env string
}

func (e *envNotFoundError) Error() string {
	return "environment not found: " + e.env
}
