package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/google/go-github/v60/github"
	"github.com/mirai-zen/forge/platform/internal/model"
	"github.com/mirai-zen/forge/platform/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
	"github.com/zeromicro/go-zero/rest/pathvar"
	"golang.org/x/oauth2"

	pb "github.com/mirai-zen/forge-proto/platform"
)

func NewCreateServiceHandler(ctx *svc.ServiceContext) *CreateServiceHandler {
	return &CreateServiceHandler{ctx: ctx}
}

type CreateServiceHandler struct {
	ctx *svc.ServiceContext
}

// createServiceBody 只解析 body 中的业务字段，project_id 从 URL 获取
type createServiceBody struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Template    string `json:"template"`
	Params      string `json:"params"`
	Creator     string `json:"creator"`
}

func (h *CreateServiceHandler) Handle(w http.ResponseWriter, r *http.Request) {
	// project_id 从 URL 路径获取
	projectID, err := strconv.ParseUint(pathvar.Vars(r)["id"], 10, 64)
	if err != nil {
		httpx.ErrorCtx(r.Context(), w, fmt.Errorf("invalid project id: %w", err))
		return
	}

	// 从 JSON body 解析 name/template/params
	var body createServiceBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.ErrorCtx(r.Context(), w, fmt.Errorf("parse body: %w", err))
		return
	}

	// 1. 查询项目
	var proj struct {
		Name    string
		GitOrg  string
		GitRepo string
	}
	if err := h.ctx.DB.Table("projects").Where("id = ?", projectID).
		Select("name, git_org, git_repo").First(&proj).Error; err != nil {
		httpx.ErrorCtx(r.Context(), w, fmt.Errorf("project not found: %w", err))
		return
	}

	// 2. 写服务元数据到数据库
	svc := &model.Service{
		ProjectID:   uint(projectID),
		Name:        body.Name,
		Description: body.Description,
		Template:    body.Template,
		ParamsJSON:  []byte(body.Params),
		Creator:     body.Creator,
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

	// 4. 直接提交服务代码到 main 分支
	repoURL, err := h.commitServiceFiles(&body, projectID, &proj)
	if err != nil {
		// 提交失败不影响服务元数据写入
		resp := &pb.CreateServiceResp{
			Id:      uint64(svc.ID),
			PrUrl: repoURL,
			Message: fmt.Sprintf("service created but commit failed: %v", err),
		}
		httpx.OkJsonCtx(r.Context(), w, resp)
		return
	}

	resp := &pb.CreateServiceResp{
		Id:      uint64(svc.ID),
		PrUrl: repoURL,
		Message: fmt.Sprintf("service %s created, committed to main", body.Name),
	}
	httpx.OkJsonCtx(r.Context(), w, resp)
}

func (h *CreateServiceHandler) commitServiceFiles(
	body *createServiceBody,
	projectID uint64,
	proj *struct {
		Name    string
		GitOrg  string
		GitRepo string
	},
) (string, error) {
	ctx := context.Background()

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: h.ctx.Config.GitHub.Token},
	)
	client := github.NewClient(oauth2.NewClient(ctx, ts))

	serviceName := body.Name
	repoPath := fmt.Sprintf("%s/%s", proj.GitOrg, proj.Name)

	files := map[string]string{
		fmt.Sprintf("%s/go.mod", serviceName): fmt.Sprintf(
			"module github.com/%s/%s\n\ngo 1.25\n", repoPath, serviceName,
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

	// 直接提交到 main 分支（自动生成代码无需 PR 审查）
	for path, content := range files {
		opts := &github.RepositoryContentFileOptions{
			Message: github.String(fmt.Sprintf("feat: add %s service", serviceName)),
			Content: []byte(content),
			Branch:  github.String("main"),
		}

		// 先检查文件是否已存在（上次失败残留）
		existing, _, _, _ := client.Repositories.GetContents(ctx, proj.GitOrg, proj.Name, path, &github.RepositoryContentGetOptions{
			Ref: "main",
		})
		if existing != nil {
			opts.SHA = existing.SHA
		}

		_, _, err := client.Repositories.CreateFile(ctx, proj.GitOrg, proj.Name, path, opts)
		if err != nil {
			return "", fmt.Errorf("create file %s: %w", path, err)
		}
	}

	repoURL := fmt.Sprintf("https://github.com/%s/%s", proj.GitOrg, proj.Name)
	return repoURL, nil
}
