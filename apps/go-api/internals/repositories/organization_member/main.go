package organization_member_repository

import (
	"fmt"
	"sort"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	yca_observe "github.com/yca-software/go-common/observer"
	yca_repository "github.com/yca-software/go-common/repository"
)

var (
	TABLE_NAME = "organization_members"
	COLUMNS    = []string{"id", "created_at", "organization_id", "user_id", "role_id"}
)

type Repository interface {
	yca_repository.Repository[models.OrganizationMember]
	Create(tx yca_repository.Tx, member *models.OrganizationMember) error
	GetByID(organizationID string, id string) (*models.OrganizationMember, error)
	GetByIDWithUser(organizationID string, id string) (*models.OrganizationMemberWithUser, error)
	GetByIDWithRole(organizationID string, id string) (*models.OrganizationMemberWithOrganizationAndRole, error)
	GetByUserIDAndOrganizationID(userID string, organizationID string) (*models.OrganizationMember, error)
	ListByUserID(userID string) (*[]models.OrganizationMemberWithOrganization, error)
	ListByUserIDWithRole(userID string) (*[]models.OrganizationMemberWithOrganizationAndRole, error)
	ListByOrganizationID(organizationID string) (*[]models.OrganizationMemberWithUser, error)
	ListUserEmailsForRole(organizationID, roleID string) ([]string, error)
	Update(tx yca_repository.Tx, member *models.OrganizationMember) error
	Delete(tx yca_repository.Tx, organizationID string, id string) error
}

type repository struct {
	yca_repository.Repository[models.OrganizationMember]
}

func New(db *sqlx.DB, metricsHook yca_observe.QueryMetricsHook) Repository {
	return &repository{
		yca_repository.NewRepository[models.OrganizationMember](db, TABLE_NAME, COLUMNS, metricsHook),
	}
}

func (r *repository) Create(tx yca_repository.Tx, member *models.OrganizationMember) error {
	return r.BaseCreate(tx, map[string]any{
		"id":              member.ID,
		"created_at":      member.CreatedAt,
		"organization_id": member.OrganizationID,
		"user_id":         member.UserID,
		"role_id":         member.RoleID,
	})
}

func (r *repository) GetByID(organizationID string, id string) (*models.OrganizationMember, error) {
	return r.BaseGet(nil, squirrel.And{
		squirrel.Eq{"id": id},
		squirrel.Eq{"organization_id": organizationID},
	}, nil)
}

func (r *repository) GetByUserIDAndOrganizationID(userID string, organizationID string) (*models.OrganizationMember, error) {
	return r.BaseGet(nil, squirrel.And{
		squirrel.Eq{"user_id": userID},
		squirrel.Eq{"organization_id": organizationID},
	}, nil)
}

func (r *repository) GetByIDWithUser(organizationID string, id string) (*models.OrganizationMemberWithUser, error) {
	columns := []string{}
	for _, column := range COLUMNS {
		columns = append(columns, fmt.Sprintf("om.%s", column))
	}
	columns = append(columns, "u.email as user_email", "u.first_name as user_first_name", "u.last_name as user_last_name")

	query := squirrel.Select(columns...).From(fmt.Sprintf("%s AS om", TABLE_NAME)).
		LeftJoin("users as u ON u.id = om.user_id").
		Where(squirrel.And{
			squirrel.Eq{"om.id": id},
			squirrel.Eq{"om.organization_id": organizationID},
		})
	sqlStr, args, err := query.PlaceholderFormat(squirrel.Dollar).ToSql()
	if err != nil {
		return nil, err
	}

	dest := new(models.OrganizationMemberWithUser)
	err = r.GetDB().Get(dest, sqlStr, args...)
	return dest, err
}

func (r *repository) GetByIDWithRole(organizationID string, id string) (*models.OrganizationMemberWithOrganizationAndRole, error) {
	columns := []string{}
	for _, column := range COLUMNS {
		columns = append(columns, fmt.Sprintf("om.%s", column))
	}
	columns = append(columns, "o.name as organization_name", "r.name as role_name", "r.permissions as role_permissions")

	query := squirrel.Select(columns...).From(fmt.Sprintf("%s AS om", TABLE_NAME)).
		LeftJoin("organizations as o ON o.id = om.organization_id").
		LeftJoin("roles as r ON r.id = om.role_id").
		Where(squirrel.And{
			squirrel.Eq{"om.id": id},
			squirrel.Eq{"om.organization_id": organizationID},
		})
	sqlStr, args, err := query.PlaceholderFormat(squirrel.Dollar).ToSql()
	if err != nil {
		return nil, err
	}

	dest := new(models.OrganizationMemberWithOrganizationAndRole)
	err = r.GetDB().Get(dest, sqlStr, args...)
	return dest, err
}

func (r *repository) ListByUserID(userID string) (*[]models.OrganizationMemberWithOrganization, error) {
	columns := []string{}
	for _, column := range COLUMNS {
		columns = append(columns, fmt.Sprintf("om.%s", column))
	}
	columns = append(columns, "o.name as organization_name")

	query := squirrel.Select(columns...).From(fmt.Sprintf("%s AS om", TABLE_NAME)).
		LeftJoin("organizations as o ON o.id = om.organization_id").
		Where(squirrel.And{
			squirrel.Eq{"om.user_id": userID},
			squirrel.Expr("o.deleted_at IS NULL"),
		})
	sqlStr, args, err := query.PlaceholderFormat(squirrel.Dollar).ToSql()
	if err != nil {
		return nil, err
	}

	dest := new([]models.OrganizationMemberWithOrganization)
	err = r.GetDB().Select(dest, sqlStr, args...)
	return dest, err
}

func (r *repository) ListByUserIDWithRole(userID string) (*[]models.OrganizationMemberWithOrganizationAndRole, error) {
	columns := []string{}
	for _, column := range COLUMNS {
		columns = append(columns, fmt.Sprintf("om.%s", column))
	}
	columns = append(columns, "o.name as organization_name", "r.name as role_name", "r.permissions as role_permissions")

	query := squirrel.Select(columns...).From(fmt.Sprintf("%s AS om", TABLE_NAME)).
		LeftJoin("organizations as o ON o.id = om.organization_id").
		LeftJoin("roles as r ON r.id = om.role_id").
		Where(squirrel.And{
			squirrel.Eq{"om.user_id": userID},
			squirrel.Expr("o.deleted_at IS NULL"),
		})
	sqlStr, args, err := query.PlaceholderFormat(squirrel.Dollar).ToSql()
	if err != nil {
		return nil, err
	}

	dest := new([]models.OrganizationMemberWithOrganizationAndRole)
	err = r.GetDB().Select(dest, sqlStr, args...)
	return dest, err
}

func (r *repository) ListByOrganizationID(organizationID string) (*[]models.OrganizationMemberWithUser, error) {
	columns := []string{}
	for _, column := range COLUMNS {
		columns = append(columns, fmt.Sprintf("om.%s", column))
	}
	columns = append(columns, "u.email as user_email", "u.first_name as user_first_name", "u.last_name as user_last_name")

	query := squirrel.Select(columns...).From(fmt.Sprintf("%s AS om", TABLE_NAME)).
		LeftJoin("users as u ON u.id = om.user_id").
		Where(squirrel.Eq{"om.organization_id": organizationID})
	sqlStr, args, err := query.PlaceholderFormat(squirrel.Dollar).ToSql()
	if err != nil {
		return nil, err
	}

	dest := new([]models.OrganizationMemberWithUser)
	err = r.GetDB().Select(dest, sqlStr, args...)
	return dest, err
}

func (r *repository) ListUserEmailsForRole(organizationID, roleID string) ([]string, error) {
	query := squirrel.Select("u.email").
		From(fmt.Sprintf("%s AS om", TABLE_NAME)).
		InnerJoin("users AS u ON u.id = om.user_id").
		Where(squirrel.And{
			squirrel.Eq{"om.organization_id": organizationID},
			squirrel.Eq{"om.role_id": roleID},
		}).
		OrderBy("u.email")
	sqlStr, args, err := query.PlaceholderFormat(squirrel.Dollar).ToSql()
	if err != nil {
		return nil, err
	}
	var emails []string
	if err := r.GetDB().Select(&emails, sqlStr, args...); err != nil {
		return nil, err
	}
	sort.Strings(emails)
	return emails, nil
}

func (r *repository) Update(tx yca_repository.Tx, member *models.OrganizationMember) error {
	return r.BaseUpdate(tx, squirrel.And{
		squirrel.Eq{"id": member.ID},
		squirrel.Eq{"organization_id": member.OrganizationID},
	}, map[string]any{
		"role_id": member.RoleID,
	})
}

func (r *repository) Delete(tx yca_repository.Tx, organizationID string, id string) error {
	return r.BaseDelete(tx, squirrel.And{
		squirrel.Eq{"id": id},
		squirrel.Eq{"organization_id": organizationID},
	})
}
