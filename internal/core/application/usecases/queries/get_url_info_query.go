package queries

import "github.com/dzhordano/urlshortener/internal/pkg/errs"

type GetURLInfoQuery struct {
	ShortURL string
}

func NewGetURLInfoQuery(shortURL string) (GetURLInfoQuery, error) {
	if shortURL == "" {
		return GetURLInfoQuery{}, errs.NewValueIsInvalidError("shortURL")
	}

	return GetURLInfoQuery{
		ShortURL: shortURL,
	}, nil
}
