package httpinbound

import (
	"net/http"

	"github.com/dzhordano/urlshortener/internal/core/application/usecases/commands"
	"github.com/dzhordano/urlshortener/internal/pkg/gen/servers"
	"github.com/labstack/echo/v4"
)

// Shorten URL and return token used for redirect
// (POST /api/v1/shorten)

func (s *Server) ShortenURL(ctx echo.Context) error {
	var req servers.ShortenURLJSONBody
	if err := ctx.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	cmd, err := commands.NewShortenURLCommand(req.Url)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	redirectToken, err := s.shortenURLCommandHandler.Handle(ctx.Request().Context(), cmd)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	}

	return ctx.JSON(http.StatusOK, echo.Map{
		"short_url": redirectToken,
	})
}
