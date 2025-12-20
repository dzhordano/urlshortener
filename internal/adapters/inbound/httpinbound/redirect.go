package httpinbound

import (
	"errors"
	"net/http"

	"github.com/dzhordano/urlshortener/internal/core/application/usecases/queries"
	"github.com/dzhordano/urlshortener/internal/pkg/errs"
	"github.com/labstack/echo/v4"
)

// Redirect to original url using provided token (using short url)
// (GET /api/v1/{token})

func (s *Server) Redirect(ctx echo.Context, token string) error {
	q, err := queries.NewRedirectQuery(token)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	resp, err := s.redirectQueryHandler.Handle(ctx.Request().Context(), q)
	if err != nil {
		if errors.Is(err, errs.ErrObjectNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "not found")
		}

		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	}

	return ctx.Redirect(http.StatusMovedPermanently, resp.OriginalURL)
}
