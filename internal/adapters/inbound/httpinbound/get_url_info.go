package httpinbound

import (
	"errors"
	"net/http"

	"github.com/dzhordano/urlshortener/internal/core/application/usecases/queries"
	"github.com/dzhordano/urlshortener/internal/pkg/errs"
	"github.com/labstack/echo/v4"
)

// Returns info about shortened url
// (GET /api/v1/{token}/info)

func (s *Server) GetShortenedURLInfo(ctx echo.Context, token string) error {
	apiKey := ctx.Request().Header.Get("X-Api-Key")
	// TODO are YOU an admin?
	if apiKey != "admin" {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	q, err := queries.NewGetURLInfoQuery(token)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	resp, err := s.getURLInfoQueryHandler.Handle(ctx.Request().Context(), q)
	if err != nil {
		if errors.Is(err, errs.ErrObjectNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "short url not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, resp)
}
