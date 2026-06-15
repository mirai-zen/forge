package logic

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/google/go-github/v60/github"
	"github.com/mirai-zen/forge/platform/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
	"github.com/zeromicro/go-zero/rest/pathvar"
	"golang.org/x/oauth2"

	"github.com/mirai-zen/forge-proto/platform"
)

func NewDeployServiceHandler(ctx *svc.ServiceContext) *DeployServiceHandler {
	return &DeployServiceHandler{ctx: ctx}
}

type DeployServiceHandler struct {
	ctx *svc.ServiceContext
}

func (h *DeployServiceHandler) Handle(w http.ResponseWriter, r *http.Request) {
	serviceID, _ := strconv.ParseUint(pathvar.Vars(r)["id"], 10, 64)

	var req platform.DeployServiceReq
	if err := httpx.Parse(r, &req); err != nil {
		httpx.Error(w, err)
		return
	}

	// 查询服务和项目信息
	var info struct {
		ServiceName string `gorm:"column:service_name"`
		ProjectName string `gorm:"column:project_name"`
		GitOrg      string `gorm:"column:git_org"`
		GitRepo     string `gorm:"column:git_repo"`
	}
	h.ctx.DB.Table("services").
		Select("services.name as service_name, projects.name as project_name, projects.git_org, projects.git_repo").
		Joins("JOIN projects ON projects.id = services.project_id").
		Where("services.id = ?", serviceID).
		First(&info)

	// 调用 GitHub Actions workflow_dispatch
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: h.ctx.Config.GitHub.Token},
	)
	client := github.NewClient(oauth2.NewClient(ctx, ts))

	workflowFile := fmt.Sprintf("deploy-%s.yaml", info.ServiceName)
	_, err := client.Actions.CreateWorkflowDispatchEventByFileName(
		ctx, info.GitOrg, info.GitRepo, workflowFile,
		github.CreateWorkflowDispatchEventRequest{
			Ref: req.Branch,
			Inputs: map[string]interface{}{
				"service":   info.ServiceName,
				"env":       req.Env,
				"branch":    req.Branch,
				"namespace": fmt.Sprintf("forge-%s", req.Env),
			},
		},
	)
	if err != nil {
		httpx.ErrorCtx(r.Context(), w, fmt.Errorf("trigger deploy: %w", err))
		return
	}

	workflowURL := fmt.Sprintf(
		"https://github.com/%s/%s/actions/workflows/%s",
		info.GitOrg, info.GitRepo, workflowFile,
	)

	resp := &platform.DeployServiceResp{
		WorkflowUrl: workflowURL,
		Message:     fmt.Sprintf("deploy %s to %s on branch %s", info.ServiceName, req.Env, req.Branch),
	}
	httpx.OkJsonCtx(r.Context(), w, resp)
}
