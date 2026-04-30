package audit_log_service

import (
	"github.com/yca-software/2chi-kit/go-api/internals/constants"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	audit_log_repository "github.com/yca-software/2chi-kit/go-api/internals/repositories/audit_log"
	yca_error "github.com/yca-software/go-common/error"
)

type ListForOrganizationRequest struct {
	OrganizationID string                                `json:"organizationId" validate:"required,uuid"`
	Filters        *audit_log_repository.AuditLogFilters `json:"filters" validate:"omitempty"`
	Limit          int                                   `json:"limit" validate:"required,min=1,max=100"`
	Offset         int                                   `json:"offset" validate:"min=0"`
}

type ListForOrganizationResponse models.PaginatedListResponse[models.AuditLogPublic]

func (s *service) ListForOrganization(req *ListForOrganizationRequest, accessInfo *models.AccessInfo) (*ListForOrganizationResponse, error) {
	if err := s.validator.ValidateStruct(req); err != nil {
		return nil, yca_error.NewUnprocessableEntityError(nil, "", &err)
	}

	org, err := s.repos.Organization.GetByID(req.OrganizationID)
	if err != nil {
		return nil, err
	}

	if err := s.authorizer.CheckOrganizationPermissionWithSubscription(accessInfo, org, constants.PERMISSION_AUDIT_READ); err != nil {
		return nil, err
	}
	if err := s.authorizer.CheckOrganizationFeature(accessInfo, org, constants.FEATURE_AUDIT_LOG); err != nil {
		return nil, err
	}

	retentionDays, _ := constants.AuditLogRetentionDaysByType[org.SubscriptionType]
	filters := req.Filters
	if retentionDays == 0 {
		return &ListForOrganizationResponse{Items: []models.AuditLogPublic{}, HasNext: false}, nil
	}
	if retentionDays > 0 {
		now := s.now()
		minStart := now.AddDate(0, 0, -retentionDays)
		if filters == nil {
			filters = &audit_log_repository.AuditLogFilters{StartDate: &minStart}
		} else if filters.StartDate == nil || filters.StartDate.Before(minStart) {
			f := *filters
			f.StartDate = &minStart
			filters = &f
		}
	}

	auditLogs, err := s.repos.AuditLog.ListByOrganizationID(req.OrganizationID, filters, req.Limit+1, req.Offset)
	if err != nil {
		return nil, err
	}

	hasNext := len(*auditLogs) > req.Limit
	if hasNext {
		items := (*auditLogs)[:req.Limit]
		auditLogs = &items
	}

	publicItems := make([]models.AuditLogPublic, 0, len(*auditLogs))
	for i := range *auditLogs {
		publicItems = append(publicItems, ToPublicAuditLog(&(*auditLogs)[i]))
	}

	return &ListForOrganizationResponse{
		Items:   publicItems,
		HasNext: hasNext,
	}, nil
}
