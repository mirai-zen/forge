package logic

import (
	"context"
	"fmt"
	"net/http"

	"github.com/mirai-zen/forge/platform/internal/model"
	"github.com/mirai-zen/forge/platform/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
	"golang.org/x/oauth2"

	"github.com/google/go-github/v60/github"
	"github.com/mirai-zen/forge-proto/platform"
)

func NewCreateProjectHandler(ctx *svc.ServiceContext) *CreateProjectHandler {
	return &CreateProjectHandler{ctx: ctx}
}

type CreateProjectHandler struct {
	ctx *svc.ServiceContext
}

func (h *CreateProjectHandler) Handle(w http.ResponseWriter, r *http.Request) {
	var req platform.CreateProjectReq
	if err := httpx.Parse(r, &req); err != nil {
		httpx.Error(w, err)
		return
	}

	// 1. 创建 GitHub 仓库
	repoURL, err := h.createGitHubRepo(&req)
	if err != nil {
		httpx.ErrorCtx(r.Context(), w, fmt.Errorf("github error: %w", err))
		return
	}

	// 2. 写数据库
	project, err := h.saveToDB(&req, repoURL)
	if err != nil {
		httpx.ErrorCtx(r.Context(), w, fmt.Errorf("db error: %w", err))
		return
	}

	resp := &platform.CreateProjectResp{
		Id:      uint64(project.ID),
		RepoUrl: repoURL,
		Message: fmt.Sprintf("Project %s created", req.Name),
	}
	httpx.OkJsonCtx(r.Context(), w, resp)
}

func (h *CreateProjectHandler) createGitHubRepo(req *platform.CreateProjectReq) (string, error) {
	ctx := context.Background()

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: h.ctx.Config.GitHub.Token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	owner := req.GitOrg
	if owner == "" {
		owner = h.ctx.Config.GitHub.Org
	}

	repo := &github.Repository{
		Name:        github.String(req.Name),
		Description: github.String(req.Description),
		Private:     github.Bool(true),
		AutoInit:    github.Bool(true),
	}

	// 先尝试作为 Organization 创建，失败则作为个人账号创建
	created, _, err := client.Repositories.Create(ctx, owner, repo)
	if err != nil {
		// 404 说明 owner 不是 Organization，尝试作为个人用户创建
		created, _, err = client.Repositories.Create(ctx, "", repo)
	}
	if err != nil {
		return "", fmt.Errorf("create repo: %w", err)
	}

	return created.GetHTMLURL(), nil
}

func (h *CreateProjectHandler) saveToDB(req *platform.CreateProjectReq, repoURL string) (*model.Project, error) {
	project := &model.Project{
		Name:     req.Name,
		GitOrg:   req.GitOrg,
		GitRepo:  repoURL,
		Template: req.Template,
	}

	if err := h.ctx.DB.Create(project).Error; err != nil {
		return nil, fmt.Errorf("insert project: %w", err)
	}

	return project, nil
}
