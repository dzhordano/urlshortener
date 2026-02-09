package httpinbound

import (
	"github.com/dzhordano/urlshortener/internal/core/application/usecases/commands"
	"github.com/dzhordano/urlshortener/internal/core/application/usecases/queries"
	"github.com/dzhordano/urlshortener/internal/pkg/errs"
	"github.com/dzhordano/urlshortener/internal/pkg/gen/servers"
)

var _ servers.ServerInterface = (*Server)(nil)

type Server struct {
	shortenURLCommandHandler commands.ShortenURLCommandHandler
	redirectQueryHandler     queries.RedirectQueryHandler
	getURLInfoQueryHandler   queries.GetURLInfoQueryHandler
}

func NewServer(
	shortenURLCommandHandler commands.ShortenURLCommandHandler,
	redirectQueryHandler queries.RedirectQueryHandler,
	getURLInfoQueryHandler queries.GetURLInfoQueryHandler,
) (*Server, error) {
	if shortenURLCommandHandler == nil {
		return nil, errs.NewValueIsRequiredError("shortenURLCommandHandler")
	}

	if redirectQueryHandler == nil {
		return nil, errs.NewValueIsRequiredError("redirectQueryHandler")
	}

	if getURLInfoQueryHandler == nil {
		return nil, errs.NewValueIsRequiredError("getURLInfoQueryHandler")
	}

	return &Server{
		shortenURLCommandHandler: shortenURLCommandHandler,
		redirectQueryHandler:     redirectQueryHandler,
		getURLInfoQueryHandler:   getURLInfoQueryHandler,
	}, nil
}
