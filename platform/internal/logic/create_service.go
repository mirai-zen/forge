package logic

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/go-github/v60/github"
	"github.com/mirai-zen/forge/platform/internal/model"
	"github.com/mirai-zen/forge/platform/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
	"golang.org/x/oauth2"

	"github.com/mirai-zen/forge-proto/platform"
)

func NewCreateServiceHandler(ctx *svc.ServiceContext) *CreateServiceHandler {
	return &CreateServiceHandler{ctx: ctx}
}

type CreateServiceHandler struct {
	ctx *svc.ServiceContext
}

func (h *CreateServiceHandler) Handle(w http.ResponseWriter, r *http.Request) {
	var req platform.CreateServiceReq
	if err := httpx.Parse(r, &req); err != nil {
		httpx.Error(w, err)
		return
	}

	// 1. 查询项目
	var proj struct {
		Name    string
		GitOrg  string
		GitRepo string
	}
	if err := h.ctx.DB.Table("projects").Where("id = ?", req.ProjectId).
		Select("name, git_org, git_repo").First(&proj).Error; err != nil {
		httpx.ErrorCtx(r.Context(), w, fmt.Errorf("project not found: %w", err))
		return
	}

	// 2. 写服务元数据到数据库
	svc := &model.Service{
		ProjectID:  uint(req.ProjectId),
		Name:       req.Name,
		Template:   req.Template,
		ParamsJSON: []byte(req.Params),
	}
	if err := h.ctx.DB.Create(svc).Error; err != nil {
		httpx.ErrorCtx(r.Context(), w, fmt.Errorf("create service: %w", err))
		return
	}

	// 3. 自动初始化 3 个环境
	envs := []model.ServiceEnv{
		{ServiceID: svc.ID, Env: "dev", Namespace: "forge-dev"},
		{ServiceID: svc.ID, Env: "staging", Namespace: "forge-staging"},
		{ServiceID: svc.ID, Env: "prod", Namespace: "forge-prod"},
	}
	h.ctx.DB.Create(&envs)

	// 4. 调用 GitHub API 提 PR
	prURL, err := h.createGitHubPR(&req, svc, proj.GitOrg, proj.Name)
	if err != nil {
		// PR 失败不影响服务写入（服务元数据已入库）
		resp := &platform.CreateServiceResp{
			Id:      uint64(svc.ID),
			PrUrl:   "",
			Message: fmt.Sprintf("service created but PR failed: %v", err),
		}
		httpx.OkJsonCtx(r.Context(), w, resp)
		return
	}

	resp := &platform.CreateServiceResp{
		Id:      uint64(svc.ID),
		PrUrl:   prURL,
		Message: fmt.Sprintf("service %s created, PR opened", req.Name),
	}
	httpx.OkJsonCtx(r.Context(), w, resp)
}

func (h *CreateServiceHandler) createGitHubPR(
	req *platform.CreateServiceReq,
	svc *model.Service,
	org, repoName string,
) (string, error) {
	ctx := context.Background()

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: h.ctx.Config.GitHub.Token},
	)
	client := github.NewClient(oauth2.NewClient(ctx, ts))

	// 生成服务代码内容（简化版：直接生成 go.mod + main.go）
	serviceName := req.Name
	repoPath := fmt.Sprintf("%s/%s", org, repoName)

	files := map[string]string{
		fmt.Sprintf("%s/go.mod", serviceName): fmt.Sprintf(
			"module github.com/%s\n\ngo 1.25\n", repoPath,
		),
		fmt.Sprintf("%s/cmd/main.go", serviceName): fmt.Sprintf(
			`package main

import "fmt"

func main() {
	fmt.Println("%s starting...")
}
`, serviceName),
		fmt.Sprintf("%s/Dockerfile", serviceName): fmt.Sprintf(
			"FROM golang:1.25-alpine AS builder\nWORKDIR /app\nCOPY go.mod ./\nRUN go mod download\nCOPY . .\nRUN CGO_ENABLED=0 go build -o /server ./cmd/\n\nFROM gcr.io/distroless/static:nonroot\nCOPY --from=builder /server /server\nUSER nonroot:nonroot\nENTRYPOINT [\"/server\"]\n",
		),
	}

	// 创建分支
	branchName := fmt.Sprintf("feat/add-%s", serviceName)
	mainRef, _, err := client.Git.GetRef(ctx, org, repoName, "refs/heads/main")
	if err != nil {
		return "", fmt.Errorf("get main ref: %w", err)
	}

	_, _, err = client.Git.CreateRef(ctx, org, repoName, &github.Reference{
		Ref: github.String("refs/heads/" + branchName),
		Object: &github.GitObject{
			SHA: mainRef.Object.SHA,
		},
	})
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		return "", fmt.Errorf("create branch: %w", err)
	}

	// 批量提交文件
	for path, content := range files {
		_, _, err = client.Repositories.CreateFile(ctx, org, repoName, path, &github.RepositoryContentFileOptions{
			Message: github.String(fmt.Sprintf("feat: add %s service", serviceName)),
			Content: []byte(content),
			Branch:  github.String(branchName),
		})
		if err != nil {
			return "", fmt.Errorf("create file %s: %w", path, err)
		}
	}

	// 创建 PR
	pr, _, err := client.PullRequests.Create(ctx, org, repoName, &github.NewPullRequest{
		Title: github.String(fmt.Sprintf("Add %s service", serviceName)),
		Head:  github.String(branchName),
		Base:  github.String("main"),
		Body: github.String(fmt.Sprintf(
			"自动生成的服务：%s\n模板：%s\n参数：%s\n\nClick Merge to start developing!",
			serviceName, req.Template, req.Params,
		)),
	})
	if err != nil {
		return "", fmt.Errorf("create PR: %w", err)
	}

	return pr.GetHTMLURL(), nil
}
