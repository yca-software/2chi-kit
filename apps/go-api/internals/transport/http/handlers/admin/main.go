package admin_handler

import (
	"github.com/labstack/echo/v4"
	"github.com/yca-software/2chi-kit/go-api/internals/services"
	admin_organization_handler "github.com/yca-software/2chi-kit/go-api/internals/transport/http/handlers/admin/organization"
	admin_user_handler "github.com/yca-software/2chi-kit/go-api/internals/transport/http/handlers/admin/user"
	http_middlewares "github.com/yca-software/2chi-kit/go-api/internals/transport/http/middlewares"
)

type Handler struct {
	services    *services.Services
	middlewares http_middlewares.Middlewares
}

func New(services *services.Services, middlewares http_middlewares.Middlewares) *Handler {
	return &Handler{
		services:    services,
		middlewares: middlewares,
	}
}

func (h *Handler) RegisterEndpoints(router *echo.Group) {
	adminOrganizationHandler := admin_organization_handler.New(h.services)
	adminUserHandler := admin_user_handler.New(h.services)

	group := router.Group("/admin", h.middlewares.PrivateRoute, h.middlewares.AdminRoute)
	adminOrganizationHandler.RegisterEndpoints(group)
	adminUserHandler.RegisterEndpoints(group)
}
