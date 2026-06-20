package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/go-github/v60/github"
	"github.com/mirai-zen/forge/platform/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
	"github.com/zeromicro/go-zero/rest/pathvar"
	"golang.org/x/oauth2"

	pb "github.com/mirai-zen/forge-proto/platform"
)

func NewDeployServiceHandler(ctx *svc.ServiceContext) *DeployServiceHandler {
	return &DeployServiceHandler{ctx: ctx}
}

type DeployServiceHandler struct {
	ctx *svc.ServiceContext
}

// deployBody 部署请求 body（env/branch），service_id 从 URL 获取
type deployBody struct {
	Env    string `json:"env"`
	Branch string `json:"branch"`
}

func (h *DeployServiceHandler) Handle(w http.ResponseWriter, r *http.Request) {
	serviceID, _ := strconv.ParseUint(pathvar.Vars(r)["id"], 10, 64)

	var body deployBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.ErrorCtx(r.Context(), w, fmt.Errorf("parse body: %w", err))
		return
	}

	// 查询服务和项目信息
	var info struct {
		ServiceName string
		ProjectName string
		GitOrg      string
		GitRepo     string
	}
	if err := h.ctx.DB.Raw(
		"SELECT s.name as service_name, p.name as project_name, p.git_org, p.git_repo FROM services s JOIN projects p ON p.id = s.project_id WHERE s.id = ?",
		serviceID,
	).Scan(&info).Error; err != nil {
		httpx.ErrorCtx(r.Context(), w, fmt.Errorf("service not found: %w", err))
		return
	}
	if info.ServiceName == "" {
		httpx.ErrorCtx(r.Context(), w, fmt.Errorf("service not found"))
		return
	}

	// 如果 git_repo 是完整 URL，提取仓库名
	repoName := info.GitRepo
	if strings.Contains(repoName, "/") && strings.Contains(repoName, ".") {
		// 从 https://github.com/org/repo 提取 repo
		parts := strings.Split(repoName, "/")
		repoName = parts[len(parts)-1]
	}

	// 调用 GitHub Actions workflow_dispatch
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: h.ctx.Config.GitHub.Token},
	)
	client := github.NewClient(oauth2.NewClient(ctx, ts))

	workflowFile := fmt.Sprintf("deploy-%s.yaml", info.ServiceName)
	_, err := client.Actions.CreateWorkflowDispatchEventByFileName(
		ctx, info.GitOrg, repoName, workflowFile,
		github.CreateWorkflowDispatchEventRequest{
			Ref: body.Branch,
			Inputs: map[string]interface{}{
				"service":   info.ServiceName,
				"env":       body.Env,
				"branch":    body.Branch,
				"namespace": fmt.Sprintf("forge-%s", body.Env),
			},
		},
	)
	if err != nil {
		httpx.ErrorCtx(r.Context(), w, fmt.Errorf("trigger deploy: %w", err))
		return
	}

	workflowURL := fmt.Sprintf(
		"https://github.com/%s/%s/actions/workflows/%s",
		info.GitOrg, repoName, workflowFile,
	)

	resp := &pb.DeployServiceResp{
		WorkflowUrl: workflowURL,
		Message:     fmt.Sprintf("deploy %s to %s on branch %s", info.ServiceName, body.Env, body.Branch),
	}
	httpx.OkJsonCtx(r.Context(), w, resp)
}
