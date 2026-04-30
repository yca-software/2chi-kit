package http_middlewares

import (
	"github.com/labstack/echo/v4"
	"github.com/yca-software/2chi-kit/go-api/internals/constants"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	yca_error "github.com/yca-software/go-common/error"
)

func (m *middlewares) AdminRoute(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		accessInfo := c.Get("accessInfo").(*models.AccessInfo)

		if accessInfo == nil {
			return yca_error.NewUnauthorizedError(nil, constants.UNAUTHORIZED_CODE, nil)
		}

		if accessInfo.User == nil || !accessInfo.User.IsAdmin {
			return yca_error.NewForbiddenError(nil, constants.FORBIDDEN_CODE, nil)
		}

		return next(c)
	}
}
