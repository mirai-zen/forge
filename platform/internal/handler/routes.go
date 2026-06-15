package handler

import (
	"net/http"

	"github.com/mirai-zen/forge/platform/internal/logic"
	"github.com/mirai-zen/forge/platform/internal/svc"
	"github.com/zeromicro/go-zero/rest"
)

func RegisterRoutes(server *rest.Server, ctx *svc.ServiceContext) {
	server.AddRoutes(
		[]rest.Route{
			{Method: http.MethodPost, Path: "/", Handler: logic.NewCreateProjectHandler(ctx).Handle},
			{Method: http.MethodGet, Path: "/", Handler: logic.NewListProjectsHandler(ctx).Handle},
		},
		rest.WithPrefix("/api/platform/projects"),
	)

	server.AddRoutes(
		[]rest.Route{
			{Method: http.MethodGet, Path: "/:id", Handler: logic.NewGetProjectHandler(ctx).Handle},
			{Method: http.MethodPost, Path: "/:id/services", Handler: logic.NewCreateServiceHandler(ctx).Handle},
		},
		rest.WithPrefix("/api/platform/projects"),
	)

	server.AddRoutes(
		[]rest.Route{
			{Method: http.MethodGet, Path: "/:id", Handler: logic.NewGetServiceHandler(ctx).Handle},
			{Method: http.MethodPost, Path: "/:id/deploy", Handler: logic.NewDeployServiceHandler(ctx).Handle},
			{Method: http.MethodGet, Path: "/:id/envs/:env", Handler: logic.NewGetEnvStatusHandler(ctx).Handle},
		},
		rest.WithPrefix("/api/platform/services"),
	)

	server.AddRoutes(
		[]rest.Route{
			{Method: http.MethodGet, Path: "/", Handler: logic.NewListTemplatesHandler(ctx).Handle},
		},
		rest.WithPrefix("/api/platform/templates"),
	)
}
