package admin_organization_handler

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	audit_log_repository "github.com/yca-software/2chi-kit/go-api/internals/repositories/audit_log"
	"github.com/yca-software/2chi-kit/go-api/internals/services"
	audit_log_service "github.com/yca-software/2chi-kit/go-api/internals/services/audit_log"
	organization_service "github.com/yca-software/2chi-kit/go-api/internals/services/organization"
	yca_error "github.com/yca-software/go-common/error"
	yca_http "github.com/yca-software/go-common/http"
)

type Handler struct {
	services *services.Services
}

func New(services *services.Services) *Handler {
	return &Handler{
		services: services,
	}
}

func (h *Handler) RegisterEndpoints(router *echo.Group) {
	group := router.Group("/organization")

	group.GET("", h.ListOrganizations)
	group.GET("/archived", h.ListArchivedOrganizations)
	group.GET("/archived/:orgId", h.GetArchivedOrganization)
	group.POST("/archived/:orgId/restore", h.RestoreOrganization)
	group.POST("", h.CreateOrganizationWithCustomSubscription)
	group.PATCH("/:orgId/subscription", h.UpdateOrganizationSubscriptionSettings)
	group.GET("/:orgId", h.GetOrganization)
	group.GET("/:orgId/audit-log", h.ListOrganizationAuditLogs)
}

// CreateOrganizationWithCustomSubscription godoc
// @Summary      Create organization with custom subscription (admin)
// @Description  Same as regular org create (Paddle customer, cleanups). Sets custom_subscription and subscription fields; sends an invitation to the owner email to join as owner. Enterprise = subscription type 3.
// @Tags         Admin Organization
// @Accept       json
// @Produce      json
// @Param        body  body  organization_service.AdminCreateOrganizationWithCustomSubscriptionRequest  true  "Request body"
// @Success      201   {object}  models.Organization
// @Failure      400   {object}  models.ErrorResponse
// @Failure      401   {object}  models.ErrorResponse
// @Failure      403   {object}  models.ErrorResponse
// @Failure      422   {object}  models.ErrorResponse
// @Failure      500   {object}  models.ErrorResponse
// @Security     BearerAuth
// @Router       /admin/organization [post]
func (h *Handler) CreateOrganizationWithCustomSubscription(c echo.Context) error {
	accessInfo := c.Get("accessInfo").(*models.AccessInfo)
	var req organization_service.AdminCreateOrganizationWithCustomSubscriptionRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	org, err := h.services.Organization.AdminCreateOrganizationWithCustomSubscription(&req, accessInfo)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, org)
}

// UpdateOrganizationSubscriptionSettings godoc
// @Summary      Update organization subscription/enterprise settings (admin)
// @Description  Updates custom subscription, subscription type, seats, and expiry. All request fields optional. Enterprise = subscription type 3.
// @Tags         Admin Organization
// @Accept       json
// @Produce      json
// @Param        orgId  path      string  true  "Organization ID"
// @Param        body   body      organization_service.AdminUpdateSubscriptionSettingsRequest  true  "Fields to update (all optional)"
// @Success      200    {object}  models.Organization
// @Failure      400    {object}  models.ErrorResponse
// @Failure      401    {object}  models.ErrorResponse
// @Failure      403    {object}  models.ErrorResponse
// @Failure      404    {object}  models.ErrorResponse
// @Failure      422    {object}  models.ErrorResponse
// @Failure      500    {object}  models.ErrorResponse
// @Security     BearerAuth
// @Router       /admin/organization/{orgId}/subscription [patch]
func (h *Handler) UpdateOrganizationSubscriptionSettings(c echo.Context) error {
	orgID := c.Param("orgId")
	accessInfo := c.Get("accessInfo").(*models.AccessInfo)
	var req organization_service.AdminUpdateSubscriptionSettingsRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	req.OrganizationID = orgID

	resp, err := h.services.Organization.AdminUpdateSubscriptionSettings(&req, accessInfo)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, resp)
}

// GetOrganization godoc
// @Summary      Get organization
// @Description  Get organization details by organization ID
// @Tags         Admin Organization
// @Accept       json
// @Produce      json
// @Param        orgId   path      string                         true  "Organization ID"
// @Success      200      {object}  models.Organization
// @Failure      400      {object}  models.ErrorResponse
// @Failure      401      {object}  models.ErrorResponse
// @Failure      403      {object}  models.ErrorResponse
// @Failure      404      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Security     BearerAuth
// @Router       /admin/organization/{orgId} [get]
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

// ListArchivedOrganizations godoc
// @Summary      List archived organizations
// @Description  Retrieves a paginated list of archived organizations (admin only). Supports search.
// @Tags         Admin Organization
// @Accept       json
// @Produce      json
// @Param        limit   query     int     false  "Number of items per page (1-100)"
// @Param        offset  query     int     false  "Page offset (0-based)"
// @Param        search  query     string  false  "Search phrase (name, address)"
// @Success      200     {object}  organization_service.PaginatedListResponse
// @Failure      401     {object}  models.ErrorResponse
// @Failure      403     {object}  models.ErrorResponse
// @Failure      500     {object}  models.ErrorResponse
// @Security     BearerAuth
// @Router       /admin/organization/archived [get]
func (h *Handler) ListArchivedOrganizations(c echo.Context) error {
	accessInfo := c.Get("accessInfo").(*models.AccessInfo)
	limit, offset := yca_http.ParseLimitOffset(c, 20, 100)
	search := c.QueryParam("search")

	resp, err := h.services.Organization.ListArchived(&organization_service.ListRequest{
		SearchPhrase: search,
		Limit:        limit,
		Offset:       offset,
	}, accessInfo)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, resp)
}

// GetArchivedOrganization godoc
// @Summary      Get archived organization
// @Description  Get archived organization details by organization ID (admin only)
// @Tags         Admin Organization
// @Accept       json
// @Produce      json
// @Param        orgId   path      string  true  "Organization ID"
// @Success      200     {object}  models.Organization
// @Failure      401     {object}  models.ErrorResponse
// @Failure      403     {object}  models.ErrorResponse
// @Failure      404     {object}  models.ErrorResponse
// @Failure      500     {object}  models.ErrorResponse
// @Security     BearerAuth
// @Router       /admin/organization/archived/{orgId} [get]
func (h *Handler) GetArchivedOrganization(c echo.Context) error {
	orgID := c.Param("orgId")
	accessInfo := c.Get("accessInfo").(*models.AccessInfo)

	resp, err := h.services.Organization.GetArchived(&organization_service.GetRequest{
		OrganizationID: orgID,
	}, accessInfo)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, resp)
}

// RestoreOrganization godoc
// @Summary      Restore archived organization
// @Description  Restores a soft-deleted organization (admin only)
// @Tags         Admin Organization
// @Accept       json
// @Produce      json
// @Param        orgId   path      string  true  "Organization ID"
// @Success      204     {object}  nil
// @Failure      401     {object}  models.ErrorResponse
// @Failure      403     {object}  models.ErrorResponse
// @Failure      404     {object}  models.ErrorResponse
// @Failure      500     {object}  models.ErrorResponse
// @Security     BearerAuth
// @Router       /admin/organization/archived/{orgId}/restore [post]
func (h *Handler) RestoreOrganization(c echo.Context) error {
	orgID := c.Param("orgId")
	accessInfo := c.Get("accessInfo").(*models.AccessInfo)

	if err := h.services.Organization.Restore(&organization_service.RestoreRequest{
		OrganizationID: orgID,
	}, accessInfo); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

// ListOrganizations godoc
// @Summary      List organizations
// @Description  Retrieves a paginated list of organizations
// @Tags         Admin Organization
// @Accept       json
// @Produce      json
// @Param        limit    query     int                            false  "Number of items per page (1-100)"
// @Param        offset   query     int                            false  "Page number (0-based)"
// @Param        search   query     string                         false  "Search phrase"
// @Success      200      {object}  organization_service.PaginatedListResponse
// @Failure      400      {object}  models.ErrorResponse
// @Failure      401      {object}  models.ErrorResponse
// @Failure      403      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Security     BearerAuth
// @Router       /admin/organization [get]
func (h *Handler) ListOrganizations(c echo.Context) error {
	accessInfo := c.Get("accessInfo").(*models.AccessInfo)
	limit, offset := yca_http.ParseLimitOffset(c, 20, 100)
	search := c.QueryParam("search")

	resp, err := h.services.Organization.List(&organization_service.ListRequest{
		SearchPhrase: search,
		Limit:        limit,
		Offset:       offset,
	}, accessInfo)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, resp)
}

// ListOrganizationAuditLogs godoc
// @Summary      List organization audit logs
// @Description  Retrieves a paginated list of audit logs for a specific organization
// @Tags         Admin Organization
// @Accept       json
// @Produce      json
// @Param        orgId   path      string                         true  "Organization ID"
// @Param        limit    query     int                            false  "Number of items per page (1-100)"
// @Param        offset   query     int                            false  "Page number (0-based)"
// @Param        startDate query    string                         false  "Start date (RFC3339)"
// @Param        endDate   query    string                         false  "End date (RFC3339)"
// @Success      200      {object}  audit_log_service.ListForOrganizationResponse
// @Failure      400      {object}  models.ErrorResponse
// @Failure      403      {object}  models.ErrorResponse
// @Failure      422      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Security     BearerAuth
// @Router       /admin/organization/{orgId}/audit-log [get]
func (h *Handler) ListOrganizationAuditLogs(c echo.Context) error {
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
