package team_member_repository

import (
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	yca_observe "github.com/yca-software/go-common/observer"
	yca_repository "github.com/yca-software/go-common/repository"
)

var (
	TABLE_NAME = "team_members"
	COLUMNS    = []string{"id", "created_at", "organization_id", "team_id", "user_id"}
)

type Repository interface {
	yca_repository.Repository[models.TeamMember]
	Create(tx yca_repository.Tx, member *models.TeamMember) error
	GetByID(organizationID string, id string) (*models.TeamMember, error)
	GetByIDWithUser(tx yca_repository.Tx, organizationID string, id string) (*models.TeamMemberWithUser, error)
	ListByUserID(userID string) (*[]models.TeamMemberWithTeam, error)
	ListByTeamID(organizationID string, teamID string) (*[]models.TeamMemberWithUser, error)
	ListByOrganizationID(organizationID string) (*[]models.TeamMemberWithUser, error)
	Delete(tx yca_repository.Tx, organizationID string, id string) error
}

type repository struct {
	yca_repository.Repository[models.TeamMember]
}

func New(db *sqlx.DB, metricsHook yca_observe.QueryMetricsHook) Repository {
	return &repository{
		yca_repository.NewRepository[models.TeamMember](db, TABLE_NAME, COLUMNS, metricsHook),
	}
}

func (r *repository) Create(tx yca_repository.Tx, member *models.TeamMember) error {
	return r.BaseCreate(tx, map[string]any{
		"id":              member.ID,
		"created_at":      member.CreatedAt,
		"organization_id": member.OrganizationID,
		"team_id":         member.TeamID,
		"user_id":         member.UserID,
	})
}

func (r *repository) GetByID(organizationID string, id string) (*models.TeamMember, error) {
	return r.BaseGet(nil, squirrel.And{
		squirrel.Eq{"id": id},
		squirrel.Eq{"organization_id": organizationID},
	}, nil)
}

func (r *repository) GetByIDWithUser(tx yca_repository.Tx, organizationID string, id string) (*models.TeamMemberWithUser, error) {
	columns := []string{}
	for _, column := range COLUMNS {
		columns = append(columns, fmt.Sprintf("tm.%s", column))
	}
	columns = append(columns, "u.email as user_email", "u.first_name as user_first_name", "u.last_name as user_last_name")

	query := squirrel.Select(columns...).From(fmt.Sprintf("%s AS tm", TABLE_NAME)).
		LeftJoin("users as u ON u.id = tm.user_id").
		Where(squirrel.And{
			squirrel.Eq{"tm.id": id},
			squirrel.Eq{"tm.organization_id": organizationID},
		})
	sqlStr, args, err := query.PlaceholderFormat(squirrel.Dollar).ToSql()
	if err != nil {
		return nil, err
	}

	dest := new(models.TeamMemberWithUser)
	if tx != nil {
		err = tx.Get(dest, sqlStr, args...)
	} else {
		err = r.GetDB().Get(dest, sqlStr, args...)
	}
	return dest, err
}

func (r *repository) ListByUserID(userID string) (*[]models.TeamMemberWithTeam, error) {
	columns := []string{}
	for _, column := range COLUMNS {
		columns = append(columns, fmt.Sprintf("tm.%s", column))
	}
	columns = append(columns, "t.name as team_name")

	query := squirrel.Select(columns...).From(fmt.Sprintf("%s AS tm", TABLE_NAME)).
		LeftJoin("teams as t ON t.id = tm.team_id").
		Where(squirrel.Eq{"tm.user_id": userID})
	sqlStr, args, err := query.PlaceholderFormat(squirrel.Dollar).ToSql()
	if err != nil {
		return nil, err
	}

	dest := new([]models.TeamMemberWithTeam)
	err = r.GetDB().Select(dest, sqlStr, args...)
	return dest, err
}

func (r *repository) ListByTeamID(organizationID string, teamID string) (*[]models.TeamMemberWithUser, error) {
	columns := []string{}
	for _, column := range COLUMNS {
		columns = append(columns, fmt.Sprintf("tm.%s", column))
	}
	columns = append(columns, "u.email as user_email", "u.first_name as user_first_name", "u.last_name as user_last_name")

	query := squirrel.Select(columns...).From(fmt.Sprintf("%s AS tm", TABLE_NAME)).
		LeftJoin("users as u ON u.id = tm.user_id").
		Where(squirrel.And{
			squirrel.Eq{"tm.organization_id": organizationID},
			squirrel.Eq{"tm.team_id": teamID},
		})
	sqlStr, args, err := query.PlaceholderFormat(squirrel.Dollar).ToSql()
	if err != nil {
		return nil, err
	}

	dest := new([]models.TeamMemberWithUser)
	err = r.GetDB().Select(dest, sqlStr, args...)
	return dest, err
}

func (r *repository) ListByOrganizationID(organizationID string) (*[]models.TeamMemberWithUser, error) {
	columns := []string{}
	for _, column := range COLUMNS {
		columns = append(columns, fmt.Sprintf("tm.%s", column))
	}
	columns = append(columns, "u.email as user_email", "u.first_name as user_first_name", "u.last_name as user_last_name")

	query := squirrel.Select(columns...).From(fmt.Sprintf("%s AS tm", TABLE_NAME)).
		LeftJoin("users as u ON u.id = tm.user_id").
		Where(squirrel.Eq{"tm.organization_id": organizationID})
	sqlStr, args, err := query.PlaceholderFormat(squirrel.Dollar).ToSql()
	if err != nil {
		return nil, err
	}

	dest := new([]models.TeamMemberWithUser)
	err = r.GetDB().Select(dest, sqlStr, args...)
	return dest, err
}

func (r *repository) Delete(tx yca_repository.Tx, organizationID string, id string) error {
	return r.BaseDelete(tx, squirrel.And{
		squirrel.Eq{"id": id},
		squirrel.Eq{"organization_id": organizationID},
	})
}
