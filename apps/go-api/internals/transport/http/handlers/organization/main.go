package organization_handler

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	audit_log_repository "github.com/yca-software/2chi-kit/go-api/internals/repositories/audit_log"
	"github.com/yca-software/2chi-kit/go-api/internals/services"
	audit_log_service "github.com/yca-software/2chi-kit/go-api/internals/services/audit_log"
	organization_service "github.com/yca-software/2chi-kit/go-api/internals/services/organization"
	paddle_service "github.com/yca-software/2chi-kit/go-api/internals/services/paddle"
	organization_api_key_handler "github.com/yca-software/2chi-kit/go-api/internals/transport/http/handlers/organization/api_key"
	organization_invitation_handler "github.com/yca-software/2chi-kit/go-api/internals/transport/http/handlers/organization/invitation"
	organization_member_handler "github.com/yca-software/2chi-kit/go-api/internals/transport/http/handlers/organization/member"
	organization_role_handler "github.com/yca-software/2chi-kit/go-api/internals/transport/http/handlers/organization/role"
	organization_team_handler "github.com/yca-software/2chi-kit/go-api/internals/transport/http/handlers/organization/team"
	http_middlewares "github.com/yca-software/2chi-kit/go-api/internals/transport/http/middlewares"
	yca_error "github.com/yca-software/go-common/error"
	yca_http "github.com/yca-software/go-common/http"
)

type Handler struct {
	services        *services.Services
	middlewares     http_middlewares.Middlewares
	rateLimitConfig *http_middlewares.RateLimitConfig
}

func New(services *services.Services, middlewares http_middlewares.Middlewares, rateLimitConfig *http_middlewares.RateLimitConfig) *Handler {
	return &Handler{services: services, middlewares: middlewares, rateLimitConfig: rateLimitConfig}
}

func (h *Handler) RegisterEndpoints(router *echo.Group) {
	apiKeyHandler := organization_api_key_handler.New(h.services)
	invitationHandler := organization_invitation_handler.New(h.services, h.rateLimitConfig)
	memberHandler := organization_member_handler.New(h.services)
	roleHandler := organization_role_handler.New(h.services)
	teamHandler := organization_team_handler.New(h.services)

	group := router.Group("/organization", h.middlewares.PrivateRoute)

	group.POST("", h.CreateOrganization)

	detailGroup := group.Group("/:orgId")
	detailGroup.GET("", h.GetOrganization)
	detailGroup.PATCH("", h.UpdateOrganization)
	detailGroup.POST("/archive", h.ArchiveOrganization)
	detailGroup.DELETE("", h.DeleteOrganization)
	detailGroup.POST("/subscription/checkout", h.CreateCheckoutSession)
	detailGroup.POST("/subscription/change-plan", h.ChangePlan)
	detailGroup.POST("/subscription/portal", h.CreateCustomerPortalSession)
	detailGroup.POST("/subscription/process-transaction", h.ProcessTransaction)
	detailGroup.GET("/audit-log", h.ListAuditLogs)

	apiKeyHandler.RegisterEndpoints(detailGroup)
	invitationHandler.RegisterEndpoints(detailGroup)
	memberHandler.RegisterEndpoints(detailGroup)
	roleHandler.RegisterEndpoints(detailGroup)
	teamHandler.RegisterEndpoints(detailGroup)
}

// CreateOrganization godoc
// @Summary      Create organization
// @Description  Creates a new organization
// @Tags         Organization
// @Accept       json
// @Produce      json
// @Param        organization body      organization_service.CreateRequest  true  "Organization request"
// @Success      201      {object}  organization_service.CreateResponse
// @Failure      400      {object}  models.ErrorResponse
// @Failure      401      {object}  models.ErrorResponse
// @Failure      403      {object}  models.ErrorResponse
// @Failure      422      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Security     BearerAuth
// @Router       /organization [post]
func (h *Handler) CreateOrganization(c echo.Context) error {
	accessInfo := c.Get("accessInfo").(*models.AccessInfo)
	var req organization_service.CreateRequest
	if err := c.Bind(&req); err != nil {
		return yca_error.NewBadRequestError(nil, "", &err)
	}

	resp, err := h.services.Organization.Create(&req, accessInfo)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, resp)
}

// DeleteOrganization godoc
// @Summary      Delete organization
// @Description  Deletes a specific organization by its ID
// @Tags         Organization
// @Accept       json
// @Produce      json
// @Param        orgId   path      string                         true  "Organization ID"
// @Success      204      {object}  nil
// @Failure      400      {object}  models.ErrorResponse
// @Failure      403      {object}  models.ErrorResponse
// @Failure      404      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Security     BearerAuth
// @Router       /organization/{orgId} [delete]
func (h *Handler) DeleteOrganization(c echo.Context) error {
	accessInfo := c.Get("accessInfo").(*models.AccessInfo)
	orgID := c.Param("orgId")

	if err := h.services.Organization.Delete(&organization_service.DeleteRequest{
		OrganizationID: orgID,
	}, accessInfo); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

// ArchiveOrganization godoc
// @Summary      Archive organization
// @Description  Archives an organization by its ID (soft delete). Requires org:delete permission (e.g. Owner).
// @Tags         Organization
// @Accept       json
// @Produce      json
// @Param        orgId   path      string  true  "Organization ID"
// @Success      204      {object}  nil
// @Failure      400      {object}  models.ErrorResponse
// @Failure      403      {object}  models.ErrorResponse
// @Failure      404      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Security     BearerAuth
// @Router       /organization/{orgId}/archive [post]
func (h *Handler) ArchiveOrganization(c echo.Context) error {
	accessInfo := c.Get("accessInfo").(*models.AccessInfo)
	orgID := c.Param("orgId")

	if err := h.services.Organization.Archive(&organization_service.ArchiveRequest{
		OrganizationID: orgID,
	}, accessInfo); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

// GetOrganization godoc
// @Summary      Get organization
// @Description  Retrieves a specific organization by its ID
// @Tags         Organization
// @Accept       json
// @Produce      json
// @Param        orgId   path      string                         true  "Organization ID"
// @Success      200      {object}  models.Organization
// @Failure      400      {object}  models.ErrorResponse
// @Failure      403      {object}  models.ErrorResponse
// @Failure      404      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Security     BearerAuth
// @Router       /organization/{orgId} [get]
func (h *Handler) GetOrganization(c echo.Context) error {
	orgID := c.Param("orgId")
	accessInfo := c.Get("accessInfo").(*models.AccessInfo)

	resp, err := h.services.Organization.Get(&organization_service.GetRequest{
		OrganizationID: orgID,
	}, accessInfo)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, resp)
}

// UpdateOrganization godoc
// @Summary      Update organization
// @Description  Updates a specific organization by its ID
// @Tags         Organization
// @Accept       json
// @Produce      json
// @Param        orgId   path      string                         true  "Organization ID"
// @Param        organization body      organization_service.UpdateRequest  true  "Organization request"
// @Success      200      {object}  models.Organization
// @Failure      400      {object}  models.ErrorResponse
// @Failure      403      {object}  models.ErrorResponse
// @Failure      404      {object}  models.ErrorResponse
// @Failure      422      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Security     BearerAuth
// @Router       /organization/{orgId} [patch]
func (h *Handler) UpdateOrganization(c echo.Context) error {
	orgID := c.Param("orgId")
	accessInfo := c.Get("accessInfo").(*models.AccessInfo)

	var req organization_service.UpdateRequest
	if err := c.Bind(&req); err != nil {
		return yca_error.NewBadRequestError(nil, "", &err)
	}
	req.OrganizationID = orgID

	resp, err := h.services.Organization.Update(&req, accessInfo)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, resp)
}

// CreateCheckoutSession godoc
// @Summary      Create checkout session
// @Description  Creates a Paddle transaction for the organization (upgrade/downgrade). Returns transaction ID for opening checkout in-app with Paddle.js.
// @Tags         Organization
// @Accept       json
// @Produce      json
// @Param        orgId   path      string  true  "Organization ID"
// @Param        body    body      paddle_service.CreateCheckoutSessionRequest  true  "Checkout request (planId = Paddle price ID)"
// @Success      200      {object}  paddle_service.CheckoutSessionResponse
// @Failure      400      {object}  models.ErrorResponse
// @Failure      403      {object}  models.ErrorResponse
// @Failure      404      {object}  models.ErrorResponse
// @Failure      422      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Security     BearerAuth
// @Router       /organization/{orgId}/subscription/checkout [post]
func (h *Handler) CreateCheckoutSession(c echo.Context) error {
	orgID := c.Param("orgId")
	accessInfo := c.Get("accessInfo").(*models.AccessInfo)

	var req paddle_service.CreateCheckoutSessionRequest
	if err := c.Bind(&req); err != nil {
		return yca_error.NewBadRequestError(nil, "", &err)
	}
	req.OrganizationID = orgID

	resp, err := h.services.Paddle.CreateCheckoutSession(req, accessInfo)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, resp)
}

// CreateCustomerPortalSession godoc
// @Summary      Create customer portal session
// @Description  Creates a Paddle customer portal session for the organization. Returns URL to redirect the user to manage subscription/payment.
// @Tags         Organization
// @Accept       json
// @Produce      json
// @Param        orgId   path      string  true  "Organization ID"
// @Success      200      {object}  paddle_service.CustomerPortalSessionResponse
// @Failure      400      {object}  models.ErrorResponse
// @Failure      403      {object}  models.ErrorResponse
// @Failure      404      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Security     BearerAuth
// @Router       /organization/{orgId}/subscription/portal [post]
func (h *Handler) CreateCustomerPortalSession(c echo.Context) error {
	orgID := c.Param("orgId")
	accessInfo := c.Get("accessInfo").(*models.AccessInfo)

	resp, err := h.services.Paddle.CreateCustomerPortalSession(paddle_service.CreateCustomerPortalSessionRequest{
		OrganizationID: orgID,
	}, accessInfo)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, resp)
}

// ChangePlan godoc
// @Summary      Change subscription plan
// @Description  Updates the organization's Paddle subscription to a new plan (upgrade/downgrade). Requires an existing subscription.
// @Tags         Organization
// @Accept       json
// @Produce      json
// @Param        orgId   path      string  true  "Organization ID"
// @Param        body    body      paddle_service.ChangePlanRequest  true  "Change plan request (planId = Paddle price ID)"
// @Success      200     {object}  paddle_service.ChangePlanResult
// @Failure      400     {object}  models.ErrorResponse
// @Failure      403     {object}  models.ErrorResponse
// @Failure      404     {object}  models.ErrorResponse
// @Failure      422     {object}  models.ErrorResponse
// @Failure      500     {object}  models.ErrorResponse
// @Security     BearerAuth
// @Router       /organization/{orgId}/subscription/change-plan [post]
func (h *Handler) ChangePlan(c echo.Context) error {
	orgID := c.Param("orgId")
	accessInfo := c.Get("accessInfo").(*models.AccessInfo)

	var req paddle_service.ChangePlanRequest
	if err := c.Bind(&req); err != nil {
		return yca_error.NewBadRequestError(nil, "", &err)
	}
	req.OrganizationID = orgID

	result, err := h.services.Paddle.ChangePlan(c.Request().Context(), &req, accessInfo)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, result)
}

// ProcessTransaction godoc
// @Summary      Process Paddle transaction
// @Description  Fetches a Paddle transaction and, when it belongs to this organization, updates subscription fields based on the used price.
// @Tags         Organization
// @Accept       json
// @Produce      json
// @Param        orgId   path      string                                   true  "Organization ID"
// @Param        body    body      paddle_service.ProcessTransactionRequest true  "Process transaction request (transactionId, priceId)"
// @Success      200     {object}  models.Organization
// @Failure      400     {object}  models.ErrorResponse
// @Failure      403     {object}  models.ErrorResponse
// @Failure      404     {object}  models.ErrorResponse
// @Failure      422     {object}  models.ErrorResponse
// @Failure      500     {object}  models.ErrorResponse
// @Security     BearerAuth
// @Router       /organization/{orgId}/subscription/process-transaction [post]
func (h *Handler) ProcessTransaction(c echo.Context) error {
	orgID := c.Param("orgId")
	accessInfo := c.Get("accessInfo").(*models.AccessInfo)

	var req paddle_service.ProcessTransactionRequest
	if err := c.Bind(&req); err != nil {
		return yca_error.NewBadRequestError(nil, "", &err)
	}
	req.OrganizationID = orgID

	org, err := h.services.Paddle.ProcessTransaction(c.Request().Context(), &req, accessInfo)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, org)
}

// ListAuditLogs godoc
// @Summary      List organization audit logs
// @Description  Paginated list of audit logs for the organization (subject to subscription retention).
// @Tags         Organization
// @Accept       json
// @Produce      json
// @Param        orgId      path      string  true   "Organization ID"
// @Param        limit      query     int     false  "Items per page (1-100)" default(50)
// @Param        offset     query     int     false  "Offset" default(0)
// @Param        startDate  query     string  false  "Start date (RFC3339)"
// @Param        endDate    query     string  false  "End date (RFC3339)"
// @Success      200       {object}  audit_log_service.ListForOrganizationResponse
// @Failure      400       {object}  models.ErrorResponse
// @Failure      403       {object}  models.ErrorResponse
// @Failure      422       {object}  models.ErrorResponse
// @Failure      500       {object}  models.ErrorResponse
// @Security     BearerAuth
// @Router       /organization/{orgId}/audit-log [get]
func (h *Handler) ListAuditLogs(c echo.Context) error {
	accessInfo := c.Get("accessInfo").(*models.AccessInfo)
	orgID := c.Param("orgId")
	limit, offset := yca_http.ParseLimitOffset(c, 50, 100)

	var filters *audit_log_repository.AuditLogFilters
	startDate := c.QueryParam("startDate")
	endDate := c.QueryParam("endDate")
	if startDate != "" || endDate != "" {
		f := audit_log_repository.AuditLogFilters{}
		if startDate != "" {
			parsedStartDate, err := time.Parse(time.RFC3339, startDate)
			if err != nil {
				return yca_error.NewUnprocessableEntityError(nil, "", &err)
			}
			f.StartDate = &parsedStartDate
		}
		if endDate != "" {
			parsedEndDate, err := time.Parse(time.RFC3339, endDate)
			if err != nil {
				return yca_error.NewUnprocessableEntityError(nil, "", &err)
			}
			f.EndDate = &parsedEndDate
		}
		filters = &f
	}

	resp, err := h.services.AuditLog.ListForOrganization(&audit_log_service.ListForOrganizationRequest{
		OrganizationID: orgID,
		Limit:          limit,
		Offset:         offset,
		Filters:        filters,
	}, accessInfo)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, resp)
}
