package logic

import (
	"net/http"
	"strconv"

	"github.com/mirai-zen/forge/platform/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
	"github.com/zeromicro/go-zero/rest/pathvar"

	"github.com/mirai-zen/forge-proto/platform"
)

func NewListProjectsHandler(ctx *svc.ServiceContext) *ListProjectsHandler {
	return &ListProjectsHandler{ctx: ctx}
}

type ListProjectsHandler struct {
	ctx *svc.ServiceContext
}

func (h *ListProjectsHandler) Handle(w http.ResponseWriter, r *http.Request) {
	// 手动解析 query 参数，避免 httpx.Parse 要求必填字段
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	keyword := r.URL.Query().Get("keyword")

	type ProjectRow struct {
		ID           uint   `gorm:"column:id"`
		Name         string `gorm:"column:name"`
		GitOrg       string `gorm:"column:git_org"`
		GitRepo      string `gorm:"column:git_repo"`
		Template     string `gorm:"column:template"`
		CreatedAt    string `gorm:"column:created_at"`
		ServiceCount int32  `gorm:"column:service_count"`
	}

	var rows []ProjectRow
	query := h.ctx.DB.Table("projects").
		Select("projects.*, COUNT(services.id) as service_count").
		Joins("LEFT JOIN services ON services.project_id = projects.id").
		Group("projects.id")

	if keyword != "" {
		query = query.Where("projects.name LIKE ?", "%"+keyword+"%")
	}

	var total int64
	query.Count(&total)

	query.Offset((page - 1) * pageSize).Limit(pageSize).Scan(&rows)

	var projects []*platform.ProjectInfo
	for _, r := range rows {
		projects = append(projects, &platform.ProjectInfo{
			Id:           uint64(r.ID),
			Name:         r.Name,
			GitOrg:       r.GitOrg,
			GitRepo:      r.GitRepo,
			Template:     r.Template,
			ServiceCount: r.ServiceCount,
			CreatedAt:    r.CreatedAt,
		})
	}

	if projects == nil {
		projects = []*platform.ProjectInfo{}
	}

	httpx.OkJson(w, &platform.ListProjectsResp{
		Projects: projects,
		Total:    int32(total),
	})
}

// ============================================================
// GetProject
// ============================================================

func NewGetProjectHandler(ctx *svc.ServiceContext) *GetProjectHandler {
	return &GetProjectHandler{ctx: ctx}
}

type GetProjectHandler struct {
	ctx *svc.ServiceContext
}

func (h *GetProjectHandler) Handle(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseUint(pathvar.Vars(r)["id"], 10, 64)

	var project struct {
		ID        uint   `gorm:"column:id"`
		Name      string `gorm:"column:name"`
		GitOrg    string `gorm:"column:git_org"`
		GitRepo   string `gorm:"column:git_repo"`
		Template  string `gorm:"column:template"`
		CreatedAt string `gorm:"column:created_at"`
	}

	if err := h.ctx.DB.Table("projects").Where("id = ?", id).First(&project).Error; err != nil {
		httpx.Error(w, err)
		return
	}

	var svcs []struct {
		ID          uint   `gorm:"column:id"`
		Name        string `gorm:"column:name"`
		Description string `gorm:"column:description"`
		Template    string `gorm:"column:template"`
		Creator     string `gorm:"column:creator"`
	}

	h.ctx.DB.Table("services").Where("project_id = ?", id).Scan(&svcs)

	var services []*platform.ServiceBrief
	for _, s := range svcs {
		services = append(services, &platform.ServiceBrief{
			Id:          uint64(s.ID),
			Name:        s.Name,
			Description: s.Description,
			Template:    s.Template,
			Creator:     s.Creator,
		})
	}
	if services == nil {
		services = []*platform.ServiceBrief{}
	}

	resp := &platform.GetProjectResp{
		Id:        uint64(project.ID),
		Name:      project.Name,
		GitOrg:    project.GitOrg,
		GitRepo:   project.GitRepo,
		Template:  project.Template,
		CreatedAt: project.CreatedAt,
		Services:  services,
	}

	httpx.OkJson(w, resp)
}
